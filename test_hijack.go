package main

import (
	"fmt"
	"net/http"
    "net/http/httptest"
    "io"
)

func main() {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Connection") == "Upgrade" {
            hj, _ := w.(http.Hijacker)
            conn, _, _ := hj.Hijack()
            conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
            conn.Close()
            return
        }
    }))
    defer ts.Close()
    
    req, _ := http.NewRequest("GET", ts.URL, nil)
    req.Header.Set("Connection", "Upgrade")
    req.Header.Set("Upgrade", "websocket")
    
    res, _ := http.DefaultClient.Do(req)
    fmt.Printf("Status: %d\n", res.StatusCode)
    
    if rwc, ok := res.Body.(io.ReadWriteCloser); ok {
        fmt.Println("Body is ReadWriteCloser!")
        rwc.Close()
    } else {
        fmt.Println("Body is NOT ReadWriteCloser")
    }
}
