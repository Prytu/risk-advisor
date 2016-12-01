package main

import (
	"github.com/Prytu/risk-advisor/podprovider"
	"github.com/Prytu/risk-advisor/proxy"
	"github.com/Prytu/risk-advisor/proxy/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/riskadvisor"
	"k8s.io/kubernetes/pkg/api"
	"net/http"
)

// read from somewhere
const realApiserverURL = "http://localhost:8080"
const riskAdvisorPort = ":9997"
const proxyRACommunicationPort = ":9998"
const proxySchedulerCommunicationPort = ":9999"

func main() {
	responseChannel := make(chan api.Binding, 1)
	errorChannel := make(chan error)
	podProvider := podprovider.New()

	raHandler := riskadvisorhandler.New(responseChannel, errorChannel, podProvider)

	proxy, err := proxy.New(realApiserverURL, podProvider, responseChannel, errorChannel)
	if err != nil {
		panic(err)
	}

	// TODO: Ugly, but will be fixed in issue #11
	riskAdvisor := riskadvisor.New("http://localhost" + proxyRACommunicationPort + "/advise")

	go http.ListenAndServe(proxyRACommunicationPort, raHandler)
	go http.ListenAndServe(riskAdvisorPort, riskAdvisor)
	http.ListenAndServe(proxySchedulerCommunicationPort, proxy)
}
