package server

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

type Response struct {
    Status int         `json:"status"`
    Msg    string      `json:"msg"`
    Data   interface{} `json:"data"`
}

type Server struct {
    dbTool               DBTool
    response             map[string]map[int]string
    currencyExchangeRate map[string]bool
    apiKey               string
    exchangeRateUrl      string
}

func (server *Server) Init() {
    server.initResponse()
    server.apiKey = "6F909B2C-91D6-4610-8CF6-5E7513DE254C"
    server.exchangeRateUrl = "https://rest.coinapi.io/v1/exchangerate/%s/%s?apikey=%s"
    server.currencyExchangeRate = make(map[string]bool)
    server.currencyExchangeRate["BTC"] = true
    server.currencyExchangeRate["ETH"] = true
    server.currencyExchangeRate["XMR"] = true
}

func (server *Server) initResponse() {
    server.response = make(map[string]map[int]string)
    server.response["register"] = make(map[int]string)
    server.response["register"][200] = "OK"
    server.response["register"][400] = "username already exists"
    server.response["register"][500] = "register fail"

    server.response["login"] = make(map[int]string)
    server.response["login"][200] = "OK"
    server.response["login"][400] = "no such a user"
    server.response["login"][401] = "bad username or password"
    server.response["login"][500] = "login fail"

    server.response["exchange"] = make(map[int]string)
    server.response["exchange"][200] = "OK"
    server.response["exchange"][400] = "currency not support"
    server.response["exchange"][500] = "get exchange rate fail"

    server.response["transfer"] = make(map[int]string)
    server.response["transfer"][200] = "OK"
    server.response["transfer"][400] = "set your address first"
    server.response["transfer"][401] = "login expire"
    server.response["transfer"][500] = "add transfer record fail"

    server.response["currency"] = make(map[int]string)
    server.response["currency"][200] = "OK"
    server.response["currency"][400] = "currency not support"
    server.response["currency"][401] = "login expire"
    server.response["currency"][500] = "change currency fail"
}

func (server *Server) checkCurrencySupport(currency string) bool {
    if server.currencyExchangeRate == nil {
        panic("call Init first")
    }
    return server.currencyExchangeRate[currency]
}

type ExchangeRateResponse struct {
    From string  `json:"asset_id_base"`
    To   string  `json:"asset_id_quote"`
    Rate float64 `json:"rate"`
}

func (server *Server) getExchangeRate(from, to string) float64 {
    response, err := http.Get(fmt.Sprintf(server.exchangeRateUrl, from, to, server.apiKey))
    if err != nil {
        return 0
    }
    data, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return 0
    }
    exchangeRateResponse := ExchangeRateResponse{}
    err = json.Unmarshal(data, &exchangeRateResponse)
    if err != nil {
        panic(err)
        return 0
    }
    return exchangeRateResponse.Rate
}

func (server *Server) buildResponse(status int, msg string, data interface{}) string {
    result, err := json.Marshal(Response{
        Status: status,
        Msg:    msg,
        Data:   data,
    })
    assertNoError(err)
    return string(result)
}

func (server *Server) Register(username, password string) string {
    user, status := server.dbTool.register(username, password)
    return server.buildResponse(status, server.response["register"][status], user)
}

func (server *Server) Login(username, password string) string {
    user, status := server.dbTool.login(username, password)
    return server.buildResponse(status, server.response["login"][status], user)
}

func (server *Server) ExchangeRate(from, to string) string {
    status := 200
    if !server.checkCurrencySupport(from) {
        status = 400
        return server.buildResponse(status, server.response["exchange"][status], nil)
    }

    rate := server.getExchangeRate(from, to)
    if rate < 0.00001 {
        status = 500
        return server.buildResponse(status, server.response["exchange"][status], nil)
    }

    return server.buildResponse(status, server.response["exchange"][status],
        struct{ Rate float64 `json:"rate"` }{
            Rate: rate,
        },
    )
}

func (server *Server) AddTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) string {
    status := server.dbTool.addTransferRecord(uid, token, currency, address, amount, fee, timestamp, receive)
    return server.buildResponse(status, server.response["transfer"][status], nil)
}

func (server *Server) ChangeCurrency(uid int, token, currency string) string {
    if ! server.currencyExchangeRate[currency] {
        status := 400
        return server.buildResponse(status, server.response["currency"][status], nil)
    }

    status := server.dbTool.changeCurrency(uid, token, currency)
    return server.buildResponse(status, server.response["currency"][status], nil)
}

func (server *Server) ListTransfer(uid int, token string) string {
    data, status := server.dbTool.listTransfer(uid, token)
    return server.buildResponse(status, server.response["currency"][status], data)
}
