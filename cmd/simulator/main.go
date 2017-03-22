package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
)

func main() {
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	ksf, err := kubeClient.New(*http.DefaultClient)
	if err != nil {
		log.WithError(err).Fatal("Failed to communicate with cluster when building kubeClient.\n")
	}

	// get state from apiserver
	clusterState, err := state.InitState(ksf)
	if err != nil {
		log.WithError(err).Fatal("Failed to fetch cluster state.")
	}

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
