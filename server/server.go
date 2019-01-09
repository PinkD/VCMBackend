package server

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
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
    server.dbTool.Init("tcp(10.33.4.33)")
    server.initResponse()
    server.apiKey = "6F909B2C-91D6-4610-8CF6-5E7513DE254C"
    server.exchangeRateUrl = "https://rest.coinapi.io/v1/exchangerate/%s/%s?apikey=%s"
    server.currencyExchangeRate = make(map[string]bool)
    server.currencyExchangeRate["BTC"] = true
    server.currencyExchangeRate["ETH"] = true
    server.currencyExchangeRate["XMR"] = true
    server.bindFunc()
}

func (server *Server) initResponse() {
    server.response = make(map[string]map[int]string)
    server.response["register"] = make(map[int]string)
    server.response["register"][200] = "OK"
    server.response["register"][400] = "username already exists"
    server.response["register"][403] = "empty username or password"
    server.response["register"][500] = "register fail"

    server.response["login"] = make(map[int]string)
    server.response["login"][200] = "OK"
    server.response["login"][400] = "no such a user"
    server.response["login"][401] = "bad username or password"
    server.response["login"][403] = "empty username or password"
    server.response["login"][500] = "login fail"

    server.response["exchange"] = make(map[int]string)
    server.response["exchange"][200] = "OK"
    server.response["exchange"][400] = "currency not support"
    server.response["exchange"][403] = "params empty or not enough"
    server.response["exchange"][500] = "get exchange rate fail"

    server.response["transfer"] = make(map[int]string)
    server.response["transfer"][200] = "OK"
    server.response["transfer"][400] = "set your address first"
    server.response["transfer"][401] = "login expire"
    server.response["transfer"][403] = "params empty or not enough"
    server.response["transfer"][500] = "add transfer record fail"

    server.response["profile"] = make(map[int]string)
    server.response["profile"][200] = "OK"
    server.response["profile"][400] = "currency not support"
    server.response["profile"][403] = "params empty or not enough"
    server.response["profile"][401] = "login expire"
    server.response["profile"][500] = "change currency fail"
}

func (server *Server) bindFunc() {
    http.HandleFunc("/register", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["register"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        username := postForm.Get("username")
        password := postForm.Get("password")
        log.Println(username)
        log.Println(password)
        if username == "" || password == "" {
            status := 403
            response := server.buildResponse(status, server.response["register"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.Register(username, password)
            writer.Write([]byte(response))
        }
    })
    http.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["login"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        username := postForm.Get("username")
        password := postForm.Get("password")
        if username == "" || password == "" {
            status := 403
            response := server.buildResponse(status, server.response["login"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.Login(username, password)
            writer.Write([]byte(response))
        }
    })
    http.HandleFunc("/exchange_rate", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["exchange_rate"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        from := postForm.Get("from")
        to := postForm.Get("to")
        if from == "" || to == "" {
            status := 403
            response := server.buildResponse(status, server.response["exchange_rate"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.ExchangeRate(from, to)
            writer.Write([]byte(response))
        }
    })
    http.HandleFunc("/change_currency", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["change_currency"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        uid, err := strconv.Atoi(postForm.Get("uid"))
        token := postForm.Get("token")
        currency := postForm.Get("currency")
        address := postForm.Get("address")
        if err != nil || token == "" || currency == "" {
            status := 403
            response := server.buildResponse(status, server.response["change_currency"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.ChangeProfile(uid, token, currency, address)
            writer.Write([]byte(response))
        }
    })
    http.HandleFunc("/add_transfer", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["add_transfer"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        uid, err := strconv.ParseInt(postForm.Get("uid"), 10, 0)
        token := postForm.Get("token")
        currency := postForm.Get("currency")
        address := postForm.Get("address")
        amount, _ := strconv.ParseFloat(postForm.Get("amount"), 64)
        fee, _ := strconv.ParseFloat(postForm.Get("fee"), 64)
        timestamp, _ := strconv.ParseInt(postForm.Get("timestamp"), 10, 64)
        receive, _ := strconv.ParseBool(postForm.Get("receive"))

        if err != nil || token == "" || currency == "" || address == "" || amount == 0 || fee == 0 || timestamp == 0 {
            //TODO: make sure float zero and zero are equal
            status := 403
            response := server.buildResponse(status, server.response["change_currency"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.AddTransferRecord(int(uid), token, currency, address, amount, fee, timestamp, receive)
            writer.Write([]byte(response))
        }
    })
    http.HandleFunc("/list_transfer", func(writer http.ResponseWriter, request *http.Request) {
        err := request.ParseForm()
        if err != nil {
            status := 400
            response := server.buildResponse(status, server.response["list_transfer"][status], nil)
            writer.Write([]byte(response))
        }
        postForm := request.PostForm
        uid, err := strconv.Atoi(postForm.Get("uid"))
        token := postForm.Get("token")
        if err != nil || token == "" {
            status := 403
            response := server.buildResponse(status, server.response["list_transfer"][status], nil)
            writer.Write([]byte(response))
        } else {
            response := server.ListTransfer(uid, token)
            writer.Write([]byte(response))
        }
    })
}

func (server *Server) checkCurrencySupport(currency string) bool {
    if server.currencyExchangeRate == nil {
        panic("call Init first")
    }
    return server.currencyExchangeRate[currency]
}

func (server *Server) Start(ip string, port uint16) {
    if server.currencyExchangeRate == nil {
        panic("call Init first")
    }
    panic(http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), nil))
}

func (server *Server) StartTLS(ip string, port uint16, certFile, keyFile string) {
    if server.currencyExchangeRate == nil {
        panic("call Init first")
    }
    panic(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", ip, port), certFile, keyFile, nil))
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
    if rate == 0 {
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

func (server *Server) ChangeProfile(uid int, token, currency, address string) string {
    if ! server.currencyExchangeRate[currency] {
        status := 400
        return server.buildResponse(status, server.response["profile"][status], nil)
    }
    status := server.dbTool.changeCurrency(uid, token, currency) //TODO: rearrange to `change profile`
    return server.buildResponse(status, server.response["profile"][status], nil)
}

func (server *Server) ListTransfer(uid int, token string) string {
    data, status := server.dbTool.listTransfer(uid, token)
    //only return 200 and 401
    return server.buildResponse(status, server.response["currency"][status], data)
}
