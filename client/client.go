package client

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"
)

type Response struct {
    //TODO: useless
    Status int         `json:"status"`
    Msg    string      `json:"msg"`
    Data   interface{} `json:"data"`
}

type Client struct {
    httpClient http.Client
    domain     string
}

var ContentTypeJson string = "application/json"

func (c Client) Init(domain string) {
    c.domain = domain
}

func (c Client) buildUrl(api string) (string) {
    if c.domain == "" {
        panic("call Init first")
    }
    return "https://" + c.domain + "/" + api
}

func (c Client) request(data interface{}, api string) ([]byte, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    buffer := bytes.Buffer{}
    buffer.Write(jsonData)
    response, err := c.httpClient.Post(c.buildUrl(api), ContentTypeJson, &buffer)
    if err != nil {
        return nil, err
    }
    result, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }
    return result, nil
}

type UserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (c Client) loginWrapper(username, password, api string) ([]byte) {
    user := UserRequest{
        Username: username,
        Password: password,
    }
    result, _ := c.request(user, api)
    //result must be nil while err not nil
    return result
}

func (c Client) Register(username, password string) ([]byte) {
    return c.loginWrapper(username, password, "register")
}

func (c Client) Login(username, password string) ([]byte) {
    return c.loginWrapper(username, password, "login")
}

type ExchangeRateRequest struct {
    From string `json:"from"`
    To   string `json:"to"`
}

func (c Client) ExchangeRate(from, to string) ([]byte) {
    exchangeRateRequest := ExchangeRateRequest{
        From: from,
        To:   to,
    }
    result, _ := c.request(exchangeRateRequest, "exchange_rate")
    return result
}

type ChangeCurrencyRequest struct {
    Uid      int    `json:"uid"`
    Token    string `json:"token"`
    Currency string `json:"currency"`
}

func (c Client) ChangeCurrency(uid int, token, currency string) ([]byte) {
    changeCurrencyRequest := ChangeCurrencyRequest{
        Uid:      uid,
        Token:    token,
        Currency: currency,
    }
    result, _ := c.request(changeCurrencyRequest, "change_currency")
    return result
}

type TransferRecordRequest struct {
    Uid       int     `json:"uid"`
    Token     string  `json:"token"`
    Currency  string  `json:"currency"`
    Address   string  `json:"address"`
    Amount    float64 `json:"amount"`
    Fee       float64 `json:"fee"`
    Timestamp int64   `json:"timestamp"`
    Receive   bool    `json:"receive"`
}

func (c Client) AddTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) ([]byte) {
    transferRecordRequest := TransferRecordRequest{
        Uid:       uid,
        Token:     token,
        Currency:  currency,
        Address:   address,
        Amount:    amount,
        Fee:       fee,
        Timestamp: timestamp,
        Receive:   receive,
    }
    result, _ := c.request(transferRecordRequest, "add_transfer")
    return result
}
