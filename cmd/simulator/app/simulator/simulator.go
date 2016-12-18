package simulator

import (
	"fmt"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/deckarep/golang-set"
	"k8s.io/kubernetes/pkg/api/v1"
	utilrand "k8s.io/kubernetes/pkg/util/rand"
	"log"
	"net/http"
)

type SchedulingResult struct {
	Result  string
	Message string
}

type CapacityResult struct {
	Capacity int64
}

type Simulator struct {
	brain            *brain.Brain
	schedulerHandler *schedulerHandler.SchedulerHandler
	eventChannel     <-chan *v1.Event

	// Map pod.Name to the result of scheduling attempt of that pod
	RequestPods map[string]*SchedulingResult

	// Set of pod.Name of pods from user's request that has not been processed yet
	PodsLeftToProcess mapset.Set
}

func New(brain *brain.Brain, schedulerCommunicationServer *schedulerHandler.SchedulerHandler, eventChannel <-chan *v1.Event) *Simulator {
	return &Simulator{
		brain:            brain,
		schedulerHandler: schedulerCommunicationServer,
		eventChannel:     eventChannel,
	}
}

func (s *Simulator) RunMultiplePodSimulation(podsToCreate, toDelete []*v1.Pod) []*SchedulingResult {
	requestPods := make(map[string]*SchedulingResult, len(podsToCreate))
	podsToProcess := mapset.NewSet()

	// Apply state mutations
	for _, pod := range podsToCreate {
		requestPods[pod.Name] = nil
		podsToProcess.Add(pod.Name)
		s.brain.AddPodToState(*pod)
	}

	// Run scheduler communication server
	go http.ListenAndServe(fmt.Sprintf(":%s", s.schedulerHandler.Port), s.schedulerHandler)

	for {
		event := <-s.eventChannel

		podName := event.InvolvedObject.Name
		schedulingResult := schedulingResultFromEvent(event)

		if _, ok := requestPods[podName]; ok {
			requestPods[podName] = schedulingResult
			podsToProcess.Remove(podName)
		} else {
			log.Printf(`
			Received pod scheduling event of a pod unrelated to request:
			podName: %s
			schedulingResult: %v
			`, podName, schedulingResult)
		}

		if podsToProcess.Cardinality() == 0 {
			break
		}
	}

	results := make([]*SchedulingResult, len(requestPods))
	i := 0
	for _, result := range requestPods {
		results[i] = result
		i++
	}

	return results
}

func (s *Simulator) RunSinglePodSimulation(podToSimulate *v1.Pod) *SchedulingResult {
	s.brain.AddPodToState(*podToSimulate)

	// Run scheduler communication server
	go http.ListenAndServe(fmt.Sprintf(":%s", s.schedulerHandler.Port), s.schedulerHandler)

	event := <-s.eventChannel

	return schedulingResultFromEvent(event)
}

func (s *Simulator) RunCapacitySimulation(podToSimulate *v1.Pod) *CapacityResult {
	capacity := int64(0)

	for {
		podToSimulate.Name = utilrand.String(20)
		s.brain.AddPodToState(*podToSimulate)

		event := <-s.eventChannel

		if event.Reason == "FailedScheduling" {
			break
		} else {
			capacity++
		}
	}

	return &CapacityResult{
		Capacity: capacity,
	}
}

func schedulingResultFromEvent(event *v1.Event) *SchedulingResult {
	result := event.Reason
	message := event.Message

	return &SchedulingResult{
		Result:  result,
		Message: message,
	}
}
