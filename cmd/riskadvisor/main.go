package main

import (
	"log"
	"net/http"

	"github.com/Prytu/risk-advisor/cmd/riskadvisor/app"
)

// read from somewhere
const riskAdvisorPort = ":9997"
const proxyRACommunicationPort = ":9998"
const proxyURL = "http://localhost" + proxyRACommunicationPort + "/advise"

func main() {
	// TODO: Ugly, but will be fixed in issue #11
	riskAdvisor := app.New(proxyURL)

	log.Printf("Starting risk-advisor with:\n\t- port: %v\n\t- proxy URL: %v", riskAdvisorPort, proxyURL)

	http.ListenAndServe(riskAdvisorPort, riskAdvisor)
}
