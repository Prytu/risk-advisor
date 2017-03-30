package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/initializer"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
)

func main() {
	raCommunicationPort := flag.String("ra-port", defaults.RACommunicationPort, "Port for communictaion with risk-advisor")
	schedulerCommunicationPort := flag.String("scheduler-port", defaults.SchedulerCommunicationPort, "Port for communication with scheduler")
	flag.Parse()

	var raHandlerFunc riskadvisorhandler.HTTPHandlerFunc

	ksf, err := kubeClient.New(*http.DefaultClient)
	if err != nil {
		errorMsg := "failed to communicate with cluster when building kubeClient"
		log.WithError(err).Error(errorMsg)

		raHandlerFunc = riskadvisorhandler.ErrorResponseHandler(fmt.Errorf("%s (%s)", errorMsg, err))
	} else {
		raHandlerFunc = initializer.Initialize(*schedulerCommunicationPort, state.InitState, ksf)
	}

	raHandler := riskadvisorhandler.New(raHandlerFunc)

	http.ListenAndServe(fmt.Sprintf(":%s", *raCommunicationPort), raHandler)
}
