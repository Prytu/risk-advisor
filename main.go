package main

import (
	"github.com/Prytu/risk-advisor/proxy"
	"net/http"
)

// read from somewhere
const realApiserverURL = "http://localhost:8080"

func main() {
	apiserverProxy, err := proxy.New(realApiserverURL, &proxy.FakePodProvider{})
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":9999", apiserverProxy)
}
