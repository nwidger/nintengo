package main

import (
	"flag"
	"net/http"
)

func main() {
	addr := ":8000"
	dir := "."
	flag.StringVar(&addr, "http", addr, "HTTP service address")
	flag.StringVar(&dir, "dir", dir, "directory to serve")
	flag.Parse()
	http.ListenAndServe(addr, http.FileServer(http.Dir(dir)))
}
