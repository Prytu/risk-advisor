package main

import (
	"flag"
	"fmt"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"k8s.io/kubernetes/pkg/api/v1"
	"log"
	"net/http"
)

func main() {
	apiserverAddress := flag.String("apiserver", defaults.ApiserverAddress, "Address on which real appisrver runs")
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	// get state from apiserver
	clusterState := state.InitState(*apiserverAddress)
	log.Print(clusterState)

	eventChannel := make(chan *v1.Event, 0)
	b := brain.New(clusterState, eventChannel)
	adviseHandler := riskadvisorhandler.NewMultiplePodAdviseHandler(b, *schedulerCommunicationPort, eventChannel)

	raHandler := riskadvisorhandler.New(adviseHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", *raCommunicationPort), raHandler)
}
