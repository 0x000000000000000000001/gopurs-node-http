package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nContent-Length: 0\r\n\r\n"))
		conn.Close()
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	res, err := http.DefaultClient.Do(req)
	fmt.Printf("Error: %v\n", err)
	if res != nil {
		fmt.Printf("StatusCode: %v\n", res.StatusCode)
		fmt.Printf("Upgrade Header: %v\n", res.Header.Get("Upgrade"))
	}
}
