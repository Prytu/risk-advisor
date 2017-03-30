package initializer

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/riskadvisorHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/simulator"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
)

// Returns HTTPHandlerFunc that will handle requests from riskadvisor.
// On initialization error it will return a function that responds with error message that will describe that error.
func Initialize(
	schedulerCommunicationPort string,
	initStateFunc state.InitStateFunc,
	ksf kubeClient.ClusterCommunicator,
) riskadvisorhandler.HTTPHandlerFunc {
	// get state from apiserver
	clusterState, err := initStateFunc(ksf)
	if err != nil {
		errorMsg := "failed to fetch cluster state"
		log.WithError(err).Error(errorMsg)

		return riskadvisorhandler.ErrorResponseHandler(fmt.Errorf("%s (%s)", errorMsg, err))
	}

	// Channel for sending scheduling results between brain and simulator
	eventChannel := make(chan *v1.Event)
	// Channel for simulation errors
	errorChannel := make(chan error)

	b := brain.New(clusterState, eventChannel)
	sh := schedulerHandler.New(b, schedulerCommunicationPort, errorChannel)

	s := simulator.New(b, sh, eventChannel, errorChannel)

	// Handler for risk-advisor requests (advise)
	return riskadvisorhandler.MultiplePodAdviseHandler(s)
}
