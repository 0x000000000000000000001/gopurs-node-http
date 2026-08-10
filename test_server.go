package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
    "strings"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Connection: %q\n", r.Header.Get("Connection"))
		fmt.Printf("Contains Upgrade? %v\n", strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade"))
	}))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	http.DefaultClient.Do(req)
}
