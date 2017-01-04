package riskadvisorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"log"
	"net/http"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/simulator"
	"github.com/Prytu/risk-advisor/pkg/model"
)

type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func NewMultiplePodAdviseHandler(b *brain.Brain, schedulerCommunicationPort string, eventChannel chan *v1.Event) HTTPHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		clusterMutations, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Starting scheduler server on port %s\n", schedulerCommunicationPort)

		schedHandler := schedulerHandler.New(b, schedulerCommunicationPort)
		go http.ListenAndServe(":"+schedulerCommunicationPort, schedHandler)

		s := simulator.New(b, schedHandler, eventChannel)

		result := s.RunMultiplePodSimulation(clusterMutations.ToCreate, clusterMutations.ToDelete)
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(resultJSON)
	}
}

// TODO: this is a copy paste of /\. Fix
func NewCapacityHandler(b *brain.Brain, schedulerCommunicationPort string, eventChannel chan *v1.Event) HTTPHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		clusterMutations, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Starting scheduler server on port %s\n", schedulerCommunicationPort)

		schedHandler := schedulerHandler.New(b, schedulerCommunicationPort)
		go http.ListenAndServe(":"+schedulerCommunicationPort, schedHandler)

		s := simulator.New(b, schedHandler, eventChannel)

		result := s.RunCapacitySimulation(clusterMutations.ToCreate[0])
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(resultJSON)
	}
}

func parseAdviseRequestBody(requestBody io.ReadCloser) (*model.SimulatorRequest, error) {
	body, err := ioutil.ReadAll(requestBody)
	if err != nil {
		errorMessage := fmt.Sprintf("Error reading request body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	var adviceRequest model.SimulatorRequest

	err = json.Unmarshal(body, &adviceRequest)
	if err != nil {
		errorMessage := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	return &adviceRequest, nil
}
