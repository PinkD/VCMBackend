package VCM

import "encoding/json"

type Response struct {
    Status int         `json:"status"`
    Msg    string      `json:"msg"`
    Data   interface{} `json:"data"`
}

type User struct {
    Uid      int     `json:"uid"`
    Token    string  `json:"token"`
    Currency string  `json:"currency"`
    Balance  float64 `json:"balance"`
    Address  string  `json:"address"`
}

type TransferRecord struct {
    From      string  `json:"from"`
    To        string  `json:"to"`
    Amount    float64 `json:"amount"`
    Fee       float64 `json:"fee"`
    Timestamp int64   `json:"timestamp"`
}

type Server struct {
    dbTool DBTool
}

//func (server *Server) buildResponse(status int, msg string, data interface{}) string {
func buildResponse(status int, msg string, data interface{}) string { //bind to server and reformat db tool's return value
    result, err := json.Marshal(Response{
        Status: status,
        Msg:    msg,
        Data:   data,
    })
    assertNoError(err)
    return string(result)
}
func (server *Server) Register(username, password string) string {
    return server.dbTool.register(username, password)
}

func (server *Server) Login(username, password string) string {
    return buildResponse(200, "OK", server.dbTool.login(username, password))
}

func (server *Server) ExchangeRate(currency, to string) string {
    //TODO: get exchange rate
    rate := 0.123
    return buildResponse(200, "OK",
        struct{ Rate float64 `json:"rate"` }{
            Rate: rate,
        },
    )
}

func (server *Server) AddTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) string {
    return server.dbTool.addTransferRecord(uid, token, currency, address, amount, fee, timestamp, receive)
}

func (server *Server) ChangeCurrency(uid int, token, currency string) string {
    return server.dbTool.changeCurrency(uid, token, currency)
}
