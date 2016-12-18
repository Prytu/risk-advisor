package simulator

import (
	"fmt"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/deckarep/golang-set"
	"k8s.io/kubernetes/pkg/api/v1"
	"net/http"
)

type schedulingResult struct {
	Result  string
	Message string
}

type Simulator struct {
	brain            *brain.Brain
	schedulerHandler *schedulerHandler.SchedulerHandler
	eventChannel     chan<- *v1.Event

	// Map pod.Name to the result of scheduling attempt of that pod
	RequestPods map[string]*schedulingResult

	// Set of pod.Name of pods from user's request that has not been processed yet
	PodsLeftToProcess mapset.Set
}

func New(brain *brain.Brain, schedulerCommunicationServer *schedulerHandler.SchedulerHandler) *Simulator {
	return &Simulator{
		brain:            brain,
		schedulerHandler: schedulerCommunicationServer,
	}
}

func (s *Simulator) RunMultiplePodSimulation(podsToCreate, toDelete []*v1.Pod) []*schedulingResult {
	requestPods := make(map[string]*schedulingResult, len(podsToCreate))
	podsToProcess := mapset.NewSet()

	// Apply state mutations
	for _, pod := range podsToCreate {
		requestPods[pod.Name] = nil
		podsToProcess.Add(pod.Name)
		s.brain.AddPodToState(*pod)
	}

	// Run scheduler communication server
	go http.ListenAndServe(fmt.Sprintf(":%s", s.schedulerHandler.Port), s.schedulerHandler)

	// wait for scheduling responses from Brain, parse

	return nil
}
