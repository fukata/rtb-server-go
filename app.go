package main

import (
    "fmt"
    "net/http"
    "strconv"
    "flag"
)

type Dsp struct {
    sleep_ms int
    status   int
    price    int
}

func handler(w http.ResponseWriter, r *http.Request) {
    params     := r.URL.Query()
    id         := params["id"][0]
    dsp_num, _ := strconv.Atoi( params["dsp"][0] )
    dsps       := make([]Dsp{}, dsp_num)
    for i := 0; i < dsp_num; i++ {
        sleep_ms, _ := strconv.Atoi( params["t"][0] )
        status, _   := strconv.Atoi( params["s"][0] )
        price       := 0
        if params["p"] != nil {
            price, _ = strconv.Atoi( params["p"][0] )
        }

        dsp := Dsp{ sleep_ms, status, price }
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
