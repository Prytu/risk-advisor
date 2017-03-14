package main

import (
	"flag"
	"fmt"
	"net/http"

	"k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/flags"
)

func main() {
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	// get state from apiserver
	clusterState := state.InitState()

	// Channel for sending scheduling results between scheduler communication server and simulator
	eventChannel := make(chan *v1.Event, 0)

	// Scheduler communication server with some logic
	b := brain.New(clusterState, eventChannel)

	// Handler for risk-advisor requests (advise, capacity)
	adviseHandler := riskadvisorhandler.NewMultiplePodAdviseHandler(b, *schedulerCommunicationPort, eventChannel)
	capacityHandler := riskadvisorhandler.NewCapacityHandler(b, *schedulerCommunicationPort, eventChannel)

	// Server for communication between risk-advisor and simulator
	raHandler := riskadvisorhandler.New(adviseHandler, capacityHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", *raCommunicationPort), raHandler)
}
