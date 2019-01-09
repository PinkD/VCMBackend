package server

import (
    "crypto/sha256"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "strconv"
    "strings"
    "time"
)

type DBTool struct {
    db sql.DB
}

func assertNoError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

type User struct {
    Uid      int     `json:"uid"`
    Token    string  `json:"token"`
    Currency string  `json:"currency"`
    Balance  float64 `json:"balance"`
    Address  string  `json:"address"`
}

func (dbTool *DBTool) Register(username, password string) (user *User, status int) {
    return dbTool.register(username, password)
}
func (dbTool *DBTool) Login(username, password string) (user *User, status int) {
    return dbTool.login(username, password)
}

func (dbTool *DBTool) Init(host string) {
    dbTool.init(host)
}

func (dbTool *DBTool) init(host string) {
    if host == "" {
        host = "unix(/tmp/mysql.sock)"
    }
    db, err := sql.Open("mysql", "vcm:vcm@"+host+"/vcm")
    assertNoError(err)
    dbTool.db = db

    //TODO: remove
    tx, err := db.Begin()
    assertNoError(err)
    tx.Exec("DROP TABLE IF EXISTS `transfer`;")
    tx.Exec("DROP TABLE IF EXISTS `user`;")
    err = tx.Commit()
    assertNoError(err)

    tx, err = db.Begin()
    createUserTableSQL := "CREATE TABLE IF NOT EXISTS `user` (" +
        "`id`           INTEGER     PRIMARY KEY     AUTO_INCREMENT," +
        "`username`     VARCHAR(16)     NOT NULL    UNIQUE KEY," +
        "`password`     TEXT(16)        NOT NULL," +
        "`token`        CHAR(64)        DEFAULT NULL," +
        "`currency`     VARCHAR(5)      DEFAULT 'BTC'," +
        "`balance`      DECIMAL         DEFAULT 0," +
        "`address`      VARCHAR(128)    DEFAULT ''" +
        ")"
    _, err = tx.Exec(createUserTableSQL)
    assertNoError(err)

    createTransferTableSQL := "CREATE TABLE IF NOT EXISTS `transfer` (" +
        "`uid`          INTEGER," +
        "`from`         TEXT    NOT NULL," +
        "`to`           TEXT    NOT NULL," +
        "`timestamp`    TIMESTAMP," +
        "`amount`       DECIMAL," +
        "`fee`          DECIMAL," +
        "FOREIGN KEY (`uid`) REFERENCES `user` (`id`)" +
        ")"
    _, err = tx.Exec(createTransferTableSQL)
    assertNoError(err)
    err = tx.Commit()
    assertNoError(err)
}

func (dbTool *DBTool) register(username, password string) (user *User, status int) {
    rows, err := dbTool.db.Query("SELECT * FROM `user` WHERE `username` = ?", username)
    defer rows.Close()
    assertNoError(err)
    if rows.Next() { //exists
        return nil, 400
    }
    tx, err := dbTool.db.Begin()
    _, err = tx.Exec("INSERT INTO `user` (`username`, `password`) VALUES (?, ?)", username, password)
    if err != nil {
        return nil, 500
    }
    tx.Commit()
    return dbTool.login(username, password)
}

func (dbTool *DBTool) login(username, password string) (user *User, status int) {
    seed := username + password + time.Now().String()
    data := sha256.Sum256([]byte(seed))
    var stringBuilder strings.Builder;
    for i := 0; i < len(data); i++ {
        stringBuilder.WriteString(strconv.FormatInt(int64(data[i]), 16))
    }
    token := stringBuilder.String()

    rows, err := dbTool.db.Query("SELECT `id`, `username`, `password`, `currency`, `balance`, `address` FROM `user` WHERE `username` = ?", username)
    defer rows.Close()
    assertNoError(err)
    if !rows.Next() { //not exists
        return nil, 400
    }
    user = new(User)
    var username1 string
    var password1 string
    err = rows.Scan(&user.Uid, &username1, &password1, &user.Currency, &user.Balance, &user.Address)
    assertNoError(err)
    if password != password1 {
        return nil, 401
    }
    tx, err := dbTool.db.Begin()
    tx.Exec("UPDATE `user` SET `token` = ?", token)
    tx.Commit()
    user.Token = token
    return user, 200
}

func (dbTool *DBTool) addTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) (status int) {
    rows, err := dbTool.db.Query("SELECT `address` FROM `user` WHERE `id` = ? AND `token` = ?", uid, token)
    defer rows.Close()
    assertNoError(err)
    if ! rows.Next() { //login expire
        return 401
    }
    var userAddress string
    err = rows.Scan(&userAddress)
    assertNoError(err)
    if userAddress == "" {
        return 400
    }
    tx, err := dbTool.db.Begin()
    if receive {
        _, err = tx.Exec("INSERT INTO `transfer` (`uid`, `from`, `to`, `timestamp`, `amount`, `fee`) VALUES (?, ?, ?, ?, ?, ?)",
            uid, address, userAddress, amount, fee, timestamp, receive)
    } else {
        _, err = tx.Exec("INSERT INTO `transfer` (`uid`, `from`, `to`, `timestamp`, `amount`, `fee`) VALUES (?, ?, ?, ?, ?, ?)",
            uid, userAddress, address, amount, fee, timestamp, receive)
    }
    tx.Commit()
    if err == nil {
        return 500
    } else {
        return 200
    }
}
func (dbTool *DBTool) changeCurrency(uid int, token, currency string) (status int) {
    rows, err := dbTool.db.Query("SELECT `currency` FROM `user` WHERE `id` = ? AND `token` = ?", uid, token)
    defer rows.Close()
    assertNoError(err)
    if ! rows.Next() { //login expire
        return 401
    }
    cols, err := rows.Columns()
    assertNoError(err)
    currentCurrency := cols[0]
    if currency != currentCurrency {
        tx, err := dbTool.db.Begin()
        assertNoError(err)
        tx.Exec("UPDATE `user` SET `currency` = ?", currency)
        tx.Commit()
    }
    return 200
}

type TransferRecord struct {
    From      string  `json:"from"`
    To        string  `json:"to"`
    Amount    float64 `json:"amount"`
    Fee       float64 `json:"fee"`
    Timestamp int64   `json:"timestamp"`
}

func (dbTool *DBTool) listTransfer(uid int, token string) (transferRecords []TransferRecord, status int) {
    rows, err := dbTool.db.Query("SELECT * FROM `user` WHERE `id` = ? AND `token` = ?", uid, token)
    defer rows.Close()
    assertNoError(err)
    if ! rows.Next() { //login expire
        return nil, 401
    }
    rows, err = dbTool.db.Query("SELECT * FROM `transfer` WHERE `uid` = ?", uid)
    defer rows.Close()
    assertNoError(err)
    transferRecord := TransferRecord{}
    for rows.Next() {
        err = rows.Scan(&transferRecord.From, &transferRecord.To, &transferRecord.Amount, &transferRecord.Fee, &transferRecord.Timestamp)
        assertNoError(err)
        transferRecords = append(transferRecords, transferRecord)
    }
    return transferRecords, 200
}
