package VCM

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
    db.Exec(createUserTableSQL)
    createTransferTableSQL := "CREATE TABLE IF NOT EXISTS `transfer` (" +
        "`uid`          INTEGER," +
        "`from`         TEXT    NOT NULL," +
        "`to`           TEXT    NOT NULL," +
        "`timestamp`    TIMESTAMP," +
        "`amount`       DECIMAL," +
        "`fee`          DECIMAL," +
        "FOREIGN KEY `uid` REFERENCE `user.id`" +
        ")"
    db.Exec(createTransferTableSQL)

}

func (dbTool *DBTool) login(username, password string) (*User) {
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
        return nil
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
    }
}

func (dbTool *DBTool) register(username, password string) string {
    rows, err := dbTool.db.Query("SELECT * FROM `user` WHERE `username` == $1", username)
    assertNoError(err)
    if rows.Next() { //exists
        return buildResponse(400, "username already exists", nil)
    }
    //TODO: begin first
    _, err = dbTool.db.Exec("INSERT INTO `user` (`username`, `password`) VALUES ($1, $2)}", username, password)
    assertNoError(err)

    user := dbTool.login(username, password)
    if user == nil {
        return buildResponse(500, "register fail", nil)
    } else {
        return buildResponse(200, "OK", user)
    }
    //TODO: commit here
}

func (dbTool *DBTool) addTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) string {
    return ""
}
func (dbTool *DBTool) changeCurrency(uid int, token, currency string) string {
    return ""
}
