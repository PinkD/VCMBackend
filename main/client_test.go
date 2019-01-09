package main

import (
    "../client"
    "encoding/json"
    "fmt"
    "testing"
    "time"
)

func assertError(err error, t *testing.T) {
    if err != nil {
        t.Error(err)
    }
}

func TestClientRegister(t *testing.T) {
    var c client.Client
    c.Init("127.0.0.1:8080")
    result := c.Register("PinkD", "Test")
    println(string(result))
}

func TestClientLogin(t *testing.T) {
    var c client.Client
    c.Init("127.0.0.1:8080")
    result := c.Login("PinkD", "Test")
    println(string(result))
}

func TestClientExchangeRate(t *testing.T) {
    var c client.Client
    c.Init("127.0.0.1:8080")
    result := c.ExchangeRate("BTC", "CNY")
    println(string(result))
}

func TestClientAddTransferRecord(t *testing.T) {
    var c client.Client
    c.Init("127.0.0.1:8080")
    result := c.Login("PinkD", "Test")
    var v map[string]interface{}
    v = make(map[string]interface{})
    err := json.Unmarshal(result, &v)
    if err != nil {
        panic(err)
    }
    data := v["data"].(map[string]interface{})
    uid := int(data["uid"].(float64))
    token := data["token"].(string)
    println(fmt.Sprintf("uid is %d", uid))
    result = c.AddTransferRecord(uid, token, "BTC", "abcdefg", 0.1, 0.001, time.Now().Unix(), true)
    println(string(result))
}
