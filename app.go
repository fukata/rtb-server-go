package main

import (
    "log"
    "fmt"
    "net/http"
    "strconv"
    "flag"
    "time"
    "io/ioutil"
    "encoding/json"
)

type RTBHTTPHandler struct {
    m *http.Handler
}

func (h *RTBHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params     := r.URL.Query()
    id         := params["id"][0]
    dsp_num, _ := strconv.Atoi( params["dsp"][0] )
    dsps       := make([]Dsp, dsp_num)

    // parse dsp parameters 
    for i := 0; i < dsp_num; i++ {
        sleep_ms, _ := strconv.Atoi( params[fmt.Sprintf("d%d_t", i)][0] )
        status, _   := strconv.Atoi( params[fmt.Sprintf("d%d_s", i)][0] )
        price       := 0
        price_key   := fmt.Sprintf("d%d_p", i)
        if params[price_key] != nil {
            price, _ = strconv.Atoi( params[price_key][0] )
        }

        dsp     := &Dsp{ i, id, sleep_ms, status, price }
        dsps[i] = *dsp
    }

    // do request to dsps
    results  := make([]Result, len(dsps))
    receiver := doRequests(dsps)

    // receive result
    result_num := 0
    for {
        result := <-receiver
        //log.Println(result)
        results[result.DspId] = result

        result_num++
        if len(results) == result_num { break }
    }
    //log.Println(results)

    win := doAuction(results)
    //log.Println(win)

    w.Header().Set("Content-Type", "application/json")
    response := Response{
        Id: win.ReqId,
        DspId: win.DspId,
        Price: win.Price,
    }
    bytes, _ := json.Marshal(response)
    json_str := string(bytes)
//    log.Println(json_str)
    fmt.Fprint(w, json_str)

}

type Dsp struct {
    DspId   int
    ReqId   string 
    SleepMs int
    Status   int
    Price    int
}

// {"id":"gZtaC8svEsDMyXYUukU3GM62gRPZm4mAT7MNINjrnTB3l47gBFbRxmeI4djNejm1","status":1,"price":1000}
type BidResponse struct {
    Id     string
    Status int
    Price  int
}

type Result struct {
    DspId  int
    ReqId  string
    Status  int
    Price   int
}

type Win struct {
    DspId  int
    ReqId  string
    Price  int
}

type Response struct {
    Id      string      `json:"id"`
    DspId   int         `json:"dsp_id"`
    Price   int         `json:"price"`
}

func makeClient() http.Client {
    transport := http.Transport{
        ResponseHeaderTimeout: time.Millisecond * 120,
    }

    client := http.Client{
        Transport: &transport,
    }

    return client
}

func doRequests(dsps []Dsp) <-chan Result {
    receiver := make(chan Result, len(dsps))

    for _, dsp := range dsps {
        go doRequest(dsp, receiver)
    }

    return receiver
}

func doRequest(dsp Dsp, receiver chan Result) {
    client := makeClient()
    url := fmt.Sprintf("http://localhost:8080/ad?id=%s&t=%d&s=%d&p=%d", dsp.ReqId, dsp.SleepMs, dsp.Status, dsp.Price)
    //log.Println(url)
    resp, err := client.Get(url)

    result := Result{}
    if err != nil {
        //log.Println("error")
        result.DspId = dsp.DspId
    } else {
        defer resp.Body.Close()
        body, _ := ioutil.ReadAll(resp.Body)
        //log.Println(string(body))
        var bid_res BidResponse
        err := json.Unmarshal(body, &bid_res)
        if err != nil {
            //log.Println("json parse error")
        }

        result.DspId  = dsp.DspId
        result.ReqId  = bid_res.Id
        result.Status = bid_res.Status
        result.Price  = bid_res.Price
    }

    receiver <- result
}

func doAuction(results []Result) Win {
    win := Win{}
    var max_result Result
    for _, result := range results {
        if result.Status != 1 { continue }

        if max_result.Price == 0 {
            max_result = result
        } else if result.Price > max_result.Price {
            max_result = result
        }
    }

    win.DspId = max_result.DspId
    win.ReqId = max_result.ReqId
    win.Price = max_result.Price

    return win
}

func handler(w http.ResponseWriter, r *http.Request) {
    params     := r.URL.Query()
    id         := params["id"][0]
    dsp_num, _ := strconv.Atoi( params["dsp"][0] )
    dsps       := make([]Dsp, dsp_num)

    // parse dsp parameters 
    for i := 0; i < dsp_num; i++ {
        sleep_ms, _ := strconv.Atoi( params[fmt.Sprintf("d%d_t", i)][0] )
        status, _   := strconv.Atoi( params[fmt.Sprintf("d%d_s", i)][0] )
        price       := 0
        price_key   := fmt.Sprintf("d%d_p", i)
        if params[price_key] != nil {
            price, _ = strconv.Atoi( params[price_key][0] )
        }

        dsp     := &Dsp{ i, id, sleep_ms, status, price }
        dsps[i] = *dsp
    }

    // do request to dsps
    results  := make([]Result, len(dsps))
    receiver := doRequests(dsps)

    // receive result
    result_num := 0
    for {
        result := <-receiver
        //log.Println(result)
        results[result.DspId] = result

        result_num++
        if len(results) == result_num { break }
    }
    //log.Println(results)

    win := doAuction(results)
    //log.Println(win)

    w.Header().Set("Content-Type", "application/json")
    response := Response{
        Id: win.ReqId,
        DspId: win.DspId,
        Price: win.Price,
    }
    bytes, _ := json.Marshal(response)
    json_str := string(bytes)
//    log.Println(json_str)
    fmt.Fprint(w, json_str)
}

func main() {
    port := flag.Int("port", 5000, "PORT")
    flag.Parse()

    rtbHandler := RTBHTTPHandler{}
    s := &http.Server{
        Addr:    fmt.Sprintf(":%d", *port),
        Handler: &rtbHandler,
        ReadTimeout: 500 * time.Millisecond,
    }

    //http.HandleFunc("/ad", handler)
    //log.Fatal( http.ListenAndServe(fmt.Sprintf(":%d", *port), nil) )
    log.Fatal( s.ListenAndServe() )
}
