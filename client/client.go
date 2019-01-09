package client

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

type Client struct {
    httpClient http.Client
    domain     string
    prefix     string
}

var ContentTypeJson = "application/x-www-form-urlencoded"

func (c *Client) Init(domain string) {
    c.domain = domain
    c.prefix = "http://"
}
func (c *Client) InitWithTLS(domain string, cert []byte) {
    c.domain = domain
    //TODO: use https
    c.prefix = "https://"
}

func (c *Client) buildUrl(api string) (string) {
    if c.domain == "" {
        panic("call Init first")
    }
    return c.prefix + c.domain + "/" + api
}

func (c *Client) request(data map[string]interface{}, api string) ([]byte, error) {
    var sb strings.Builder
    for k, v := range data {
        sb.WriteString(fmt.Sprintf("%s=%v&", k, v))
    }
    request := sb.String()
    request = request[:len(request)-1] //cut the last `&`
    log.Println(request)
    buffer := bytes.Buffer{}
    buffer.Write([]byte(request))
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

func (c *Client) loginWrapper(username, password, api string) ([]byte) {
    data := make(map[string]interface{})
    data["username"] = username
    data["password"] = password
    result, _ := c.request(data, api)
    //result must be nil while err not nil
    return result
}

func (c *Client) Register(username, password string) ([]byte) {
    return c.loginWrapper(username, password, "register")
}

func (c *Client) Login(username, password string) ([]byte) {
    return c.loginWrapper(username, password, "login")
}

func (c *Client) ExchangeRate(from, to string) ([]byte) {
    data := make(map[string]interface{})
    data["from"] = from
    data["to"] = to
    result, _ := c.request(data, "exchange_rate")
    return result
}

func (c *Client) ChangeCurrency(uid int, token, currency string) ([]byte) {
    data := make(map[string]interface{})
    data["uid"] = uid
    data["token"] = token
    data["currency"] = currency
    result, _ := c.request(data, "change_currency")
    return result
}

func (c *Client) AddTransferRecord(uid int, token, currency, address string, amount, fee float64, timestamp int64, receive bool) ([]byte) {
    data := make(map[string]interface{})
    data["uid"] = uid
    data["token"] = token
    data["currency"] = currency
    data["address"] = address
    data["amount"] = amount
    data["fee"] = fee
    data["timestamp"] = timestamp
    data["receive"] = receive
    result, _ := c.request(data, "add_transfer")
    return result
}

func (c *Client) listTransfer(uid int, token string) ([]byte) {
    data := make(map[string]interface{})
    data["uid"] = uid
    data["token"] = token
    result, _ := c.request(data, "list_transfer")
    return result
}
