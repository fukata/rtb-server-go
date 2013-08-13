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
    "runtime"
)

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
        MaxIdleConnsPerHost: 200,
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

var client = makeClient()
func doRequest(dsp Dsp, receiver chan Result) {
    url := fmt.Sprintf("http://dsp/ad?id=%s&t=%d&s=%d&p=%d", dsp.ReqId, dsp.SleepMs, dsp.Status, dsp.Price)
    //log.Println(url)
    resp, err := client.Get(url)

    result := Result{}
    result.DspId = dsp.DspId
    if err != nil {
        //log.Println("error")
    } else if resp != nil {
        defer resp.Body.Close()

        body, errRead := ioutil.ReadAll(resp.Body)
        if errRead != nil {
            log.Println("response read error")
        } else {
            //log.Println(string(body))
            var bidResp BidResponse
            errJson := json.Unmarshal(body, &bidResp)
            if errJson != nil {
                log.Println("json parse error")
            }

            result.ReqId  = bidResp.Id
            result.Status = bidResp.Status
            result.Price  = bidResp.Price
        }
    }

    receiver <- result
}

func doAuction(results []Result) Win {
    win := Win{}
    win.DspId = -1

    var maxResult Result
    for _, result := range results {
        if result.Status != 1 { continue }

        if maxResult.Price == 0 {
            maxResult = result
        } else if result.Price > maxResult.Price {
            maxResult = result
        }
    }

    win.DspId = maxResult.DspId
    win.ReqId = maxResult.ReqId
    win.Price = maxResult.Price

    return win
}

func handler(w http.ResponseWriter, r *http.Request) {
    params     := r.URL.Query()
    id         := params["id"][0]
    dspNum, _ := strconv.Atoi( params["dsp"][0] )
    dsps       := make([]Dsp, dspNum)

    // parse dsp parameters 
    for i := 0; i < dspNum; i++ {
        sleepMs, _ := strconv.Atoi( params[fmt.Sprintf("d%d_t", i)][0] )
        status, _   := strconv.Atoi( params[fmt.Sprintf("d%d_s", i)][0] )
        price       := 0
        price_key   := fmt.Sprintf("d%d_p", i)
        if params[price_key] != nil {
            price, _ = strconv.Atoi( params[price_key][0] )
        }

        dsp     := &Dsp{ i, id, sleepMs, status, price }
        dsps[i] = *dsp
    }

    // do request to dsps
    results  := make([]Result, len(dsps))
    receiver := doRequests(dsps)

    // receive result
    resultNum := 0
    for {
        result := <-receiver
        //log.Println(result)
        results[result.DspId] = result

        resultNum++
        if len(results) == resultNum { break }
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
    jsonStr := string(bytes)
//    log.Println(jsonStr)
    fmt.Fprint(w, jsonStr)
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    port := flag.Int("port", 5000, "PORT")
    flag.Parse()

    http.HandleFunc("/ad", handler)
    log.Fatal( http.ListenAndServe(fmt.Sprintf(":%d", *port), nil) )
}
