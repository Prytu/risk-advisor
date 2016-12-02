package main

import (
	"log"
	"net/http"

	"k8s.io/kubernetes/pkg/api"

	"github.com/Prytu/risk-advisor/cmd/proxy/app"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/podprovider"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/riskadvisorHandler"
)

// read from somewhere
const realApiserverURL = "http://localhost:8080"
const proxyRACommunicationPort = ":9998"
const proxySchedulerCommunicationPort = ":9999"

func main() {
	responseChannel := make(chan api.Binding, 1)
	errorChannel := make(chan error)
	podProvider := podprovider.New()

	raHandler := riskadvisorhandler.New(responseChannel, errorChannel, podProvider)

	proxy, err := app.New(realApiserverURL, podProvider, responseChannel, errorChannel)
	if err != nil {
		panic(err)
	}

	log.Printf("Staring proxy with:\n\t- real apiserver URL: %v\n\t- scheduler communication port: %v"+
		"\n\t- risk-advisor communication port: %v\n", realApiserverURL, proxySchedulerCommunicationPort,
		proxyRACommunicationPort)

	go http.ListenAndServe(proxyRACommunicationPort, raHandler)
	http.ListenAndServe(proxySchedulerCommunicationPort, proxy)
}
