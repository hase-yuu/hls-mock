package main

import (
	"github.com/hase-yuu/hls-mock/web"
	"net/http"
)

func main() {
	http.ListenAndServe("localhost:8000", web.New())
}
