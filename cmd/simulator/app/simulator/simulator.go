package simulator

import (
	"fmt"
	"log"
	"net/http"

	"github.com/deckarep/golang-set"
	"k8s.io/client-go/1.5/pkg/api/v1"
	utilrand "k8s.io/kubernetes/pkg/util/rand"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/Prytu/risk-advisor/pkg/model"
)

type SimulationRunner interface {
	RunMultiplePodSimulation(podsToCreate, toDelete []*v1.Pod) ([]*model.SchedulingResult, error)
}

type Simulator struct {
	brain            *brain.Brain
	schedulerHandler *schedulerHandler.SchedulerHandler
	eventChannel     <-chan *v1.Event
	errorChannel     <-chan error

	// Map pod.Name to the result of scheduling attempt of that pod
	RequestPods map[string]*model.SchedulingResult

	// Set of pod.Name of pods from user's request that has not been processed yet
	PodsLeftToProcess mapset.Set
}

func New(brain *brain.Brain, schedulerCommunicationServer *schedulerHandler.SchedulerHandler,
	eventChannel <-chan *v1.Event, errorChannel <-chan error) SimulationRunner {
	return &Simulator{
		brain:            brain,
		schedulerHandler: schedulerCommunicationServer,
		eventChannel:     eventChannel,
		errorChannel:     errorChannel,
	}
}

func (s *Simulator) RunMultiplePodSimulation(podsToCreate, toDelete []*v1.Pod) ([]*model.SchedulingResult, error) {
	requestPods := make(map[string]*model.SchedulingResult, len(podsToCreate))
	podsToProcess := mapset.NewSet()

	// Apply state mutations
	for _, pod := range podsToCreate {
		if pod.Name == "" {
			pod.Name = utilrand.String(model.MaxNameLength)
		}
		requestPods[pod.Name] = nil
		podsToProcess.Add(pod.Name)
		s.brain.AddPodToState(*pod)
	}

	// Run scheduler communication server
	log.Printf("Starting scheduler server on port %s\n", s.schedulerHandler.Port)
	go http.ListenAndServe(fmt.Sprintf(":%s", s.schedulerHandler.Port), s.schedulerHandler)

L:
	for {
		select {
		case event := <-s.eventChannel:
			podName := event.InvolvedObject.Name
			schedulingResult := schedulingResultFromEvent(event)

			if _, ok := requestPods[podName]; ok {
				requestPods[podName] = schedulingResult
				podsToProcess.Remove(podName)
			} else {
				log.Printf(`
			Received pod scheduling event of a pod unrelated to request:
			podName: %s
			schedulingResult: %v`, podName, schedulingResult)
			}

			if podsToProcess.Cardinality() == 0 {
				break L
			}
		case err := <-s.errorChannel:
			return nil, err
		}
	}

	results := make([]*model.SchedulingResult, len(requestPods))
	i := 0
	for _, result := range requestPods {
		results[i] = result
		i++
	}

	return results, nil
}

func schedulingResultFromEvent(event *v1.Event) *model.SchedulingResult {
	result := event.Reason
	message := event.Message
	podName := event.InvolvedObject.Name

	return &model.SchedulingResult{
		PodName: podName,
		Result:  result,
		Message: message,
	}
}
