package main

import (
    "../server"
    "fmt"
    "testing"
)

func TestServer(t *testing.T) {
    var s server.Server
    fmt.Println("Start listening...")
    s.Init()
    //s.Start("127.0.0.1", 80)
    s.Start("0.0.0.0", 80)
}
