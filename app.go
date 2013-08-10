package main

import (
    "fmt"
    "net/http"
    "strconv"
    "flag"
    "time"
)

type Dsp struct {
    sleep_ms int
    status   int
    price    int
}

func doRequest(Dsp *dsp) {
    time.Sleep( time.Second * 2 )
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

        dsp     := Dsp{ sleep_ms, status, price }
        dsps[i] = dsp
    }

    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, "{\"id\":\"%s\"}", id)
}

func main() {
    port := flag.Int("port", 5000, "PORT")
    flag.Parse()

    http.HandleFunc("/ad", handler)
    http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
