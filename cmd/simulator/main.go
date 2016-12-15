package main

import (
	"log"
	"flag"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/flags"
)

func main() {
	// read addresses
	apiserverAddress := flag.String("apiserver", defaults.ApiserverAddress, "Address on which real appisrver runs")
	//raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	//schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	//flag.Parse()

	// get state from apiserver
	clusterState := state.InitState(*apiserverAddress)
	log.Print(clusterState)
}
