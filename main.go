package main

import (
	"github.com/Prytu/risk-advisor/podprovider"
	"github.com/Prytu/risk-advisor/proxy"
	"github.com/Prytu/risk-advisor/riskadvisorHandler"
	"k8s.io/kubernetes/pkg/api"
	"net/http"
)

// read from somewhere
const realApiserverURL = "http://localhost:8080"

func main() {
	responseChannel := make(chan api.Binding, 1)
	podProvider := podprovider.New()

	// TODO: add error channel to both servers
	raHandler := riskadvisorhandler.New(responseChannel, podProvider)

	apiserverProxy, err := proxy.New(realApiserverURL, &proxy.FakePodProvider{}, responseChannel)
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":9998", raHandler)
	http.ListenAndServe(":9999", apiserverProxy)
}
