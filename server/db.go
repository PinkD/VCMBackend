package server

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "log"
    "strconv"
    "strings"
    "time"
)

type DBTool struct {
    db *sql.DB
}

func assertNoError(err error) {
    if err == nil {
        log.Fatal(err)
    }
}

func (dbTool *DBTool) init(host string) {
    var ctx context.Context

    host = "unix(/tmp/mysql.sock)"
    db, err := sql.Open("mysql", "vcm:vcm@"+host+"/vcm")
    assertNoError(err)
    assertNoError(err)
    db.PingContext(ctx)
    dbTool.db = db
    //TODO: test
    createUserTableSQL := "CREATE TABLE IF NOT EXISTS `user` (" +
        "`id`           INTEGER PRIMARY KEY AUTO INCREMENT," +
        "`username`     TEXT    NOT NULL    UNIQUE KEY," +
        "`password`     TEXT    NOT NULL," +
        "`token`        TEXT," +
        "`currency`     TEXT    DEFAULT BTC," +
        "`balance`      DECIMAL DEFAULT 0," +
        "`address`      TEXT    DEFAULT NULL" +
        ")"
    _, err = db.Exec(createUserTableSQL)
    assertNoError(err)

    createTransferTableSQL := "CREATE TABLE IF NOT EXISTS `transfer` (" +
        "`uid`          INTEGER," +
        "`from`         TEXT    NOT NULL," +
        "`to`           TEXT    NOT NULL," +
        "`timestamp`    TIMESTAMP," +
        "`amount`       DECIMAL," +
        "`fee`          DECIMAL," +
        "FOREIGN KEY `uid` REFERENCE `user.id`" +
        ")"
    _, err = db.Exec(createTransferTableSQL)
    assertNoError(err)

}

func (dbTool *DBTool) login(username, password string) (user *User, status int) {
    seed := username + password + time.Now().String()
    data := sha256.Sum256([]byte(seed))
    var stringBuilder strings.Builder;
    for i := range data {
        stringBuilder.WriteString(strconv.FormatInt(int64(i), 16))
    }
    token := stringBuilder.String()
    println(token)

    rows, err := dbTool.db.Query("SELECT * FROM `user` WHERE `username` == $1", username)
    assertNoError(err)
    if !rows.Next() { //exists
        return nil, 400
    }
    cols, err := rows.Columns()
    assertNoError(err)
    println(cols)
    //TODO: check
    uidStr := cols[0]
    currency := cols[1]
    balanceStr := cols[2]
    address := cols[3]
    println(uidStr)
    uid, err := strconv.Atoi(uidStr)
    assertNoError(err)
    balance, err := strconv.ParseFloat(balanceStr, 64)
    assertNoError(err)
    return &User{
        Uid:      uid,
        Token:    token,
        Currency: currency,
        Balance:  balance,
        Address:  address,
    }, 200
}

func (dbTool *DBTool) register(username, password string) (user *User, status int) {
    rows, err := dbTool.db.Query("SELECT * FROM `user` WHERE `username` == $1", username)
    assertNoError(err)
    if ! rows.Next() { //not exists
        return nil, 400
    }
    //TODO: begin first
    _, err = dbTool.db.Exec("INSERT INTO `user` (`username`, `password`) VALUES ($1, $2)}", username, password)
    if err == nil {
        return nil, 500
    }
    //TODO: commit here
    return dbTool.login(username, password)
}

func (dbTool *DBTool) addTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) (status int) {
    rows, err := dbTool.db.Query("SELECT `address` FROM `user` WHERE `uid` = $1 AND `token` = $2", uid, token)
    assertNoError(err)
    if ! rows.Next() { //login expire
        return 401
    }
    cols, err := rows.Columns()
    assertNoError(err)
    from := cols[0]
    if from == "" {
        return 400
    }
    _, err = dbTool.db.Exec("INSERT INTO `transfer` (`uid`, `from`, `to`, `timestamp`, `amount`, `fee`, `receive`) VALUES ($1, $2, $3, $4, $5, $6)",
        uid, from, address, amount, fee, timestamp, receive)
    if err == nil {
        return 500
    } else {
        return 200
    }
}
func (dbTool *DBTool) changeCurrency(uid int, token, currency string) (status int) {
    rows, err := dbTool.db.Query("SELECT `currency` FROM `user` WHERE `uid` = $1 AND `token` = $2", uid, token)
    assertNoError(err)
    if ! rows.Next() { //login expire
        return 401
    }
    cols, err := rows.Columns()
    assertNoError(err)
    currentCurrency := cols[0]
    if currency != currentCurrency {
        dbTool.db.Exec("UPDATE `user` SET `currency` = $1", currency)
    }
    return 200
}
