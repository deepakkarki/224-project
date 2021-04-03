package main

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "net/http"
)

func main() {
    // associate URLs to request handler functions
    http.HandleFunc("/", getRequest)
    http.HandleFunc("/genData", genData)

    // start web server
    log.Println("Listening on http://0.0.0.0:9999/")
    log.Fatal(http.ListenAndServe("0.0.0.0:9999", nil))
}

// handler for GET /numBytes request
func genData(w http.ResponseWriter, r *http.Request){

    // extract the query params
    bytesReq, exists := r.URL.Query()["numBytes"]

    // does numBytes exist?
    if !exists || len(bytesReq[0]) < 1 {
        log.Println("Query parameter 'numBytes' does not exist")
        return
    }

    // is it a valid number?
    numBytes, err := strconv.Atoi(bytesReq[0])
    if err != nil {
        log.Println("Invalid query parameter")
        return
    }

    // build & send a string of numBytes characters
    log.Printf("Sending %i bytes\n", numBytes)
    response := strings.Repeat("x", numBytes)
    fmt.Fprint(w, response)
}

// handler for GET / request
func getRequest(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Index")
    return
}
