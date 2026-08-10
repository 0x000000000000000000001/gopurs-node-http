package main

import (
	"fmt"
	"net/http"
)

func main() {
    fmt.Println(http.Header{"Connection": []string{"Upgrade"}}.Get("Connection") == "Upgrade")
}
