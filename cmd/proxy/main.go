package main

import (
	"log"
	"net/http"

	"github.com/Prytu/risk-advisor/cmd/proxy/app"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/podprovider"
	"github.com/Prytu/risk-advisor/cmd/proxy/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/pkg/flags"
	flag "github.com/spf13/pflag"
)

func main() {
	apiserverAddress := flag.String("apiserver", defaults.ApiserverAddress, "Address on which real appisrver runs")
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	responseChannel := make(chan interface{}, 1)
	podProvider := podprovider.New()

	raHandler := riskadvisorhandler.New(responseChannel, podProvider)

	proxy, err := app.New(*apiserverAddress, podProvider, responseChannel)
	if err != nil {
		panic(err)
	}

	log.Printf("Staring proxy with:\n\t- real apiserver URL: %v\n\t- scheduler communication port: %v"+
		"\n\t- risk-advisor communication port: %v\n", *apiserverAddress, *schedulerCommunicationPort,
		*raCommunicationPort)

	go http.ListenAndServe(":"+*raCommunicationPort, raHandler)
	http.ListenAndServe(":"+*schedulerCommunicationPort, proxy)
}
