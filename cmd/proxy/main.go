package main

import (
	"log"
	"net/http"

	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/api"

	"github.com/Prytu/risk-advisor/cmd/proxy/app"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/podprovider"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/pkg/flags"
)

func main() {
	apiserverAddress := flag.String("apiserver", defaults.ApiserverAddress, "Address on which real appisrver runs")
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	responseChannel := make(chan api.Binding, 1)
	errorChannel := make(chan error)
	podProvider := podprovider.New()

	raHandler := riskadvisorhandler.New(responseChannel, errorChannel, podProvider)

	proxy, err := app.New(*apiserverAddress, podProvider, responseChannel, errorChannel)
	if err != nil {
		panic(err)
	}

	log.Printf("Staring proxy with:\n\t- real apiserver URL: %v\n\t- scheduler communication port: %v"+
		"\n\t- risk-advisor communication port: %v\n", *apiserverAddress, *schedulerCommunicationPort,
		*raCommunicationPort)

	go http.ListenAndServe(":"+*raCommunicationPort, raHandler)
	http.ListenAndServe(":"+*schedulerCommunicationPort, proxy)
}
