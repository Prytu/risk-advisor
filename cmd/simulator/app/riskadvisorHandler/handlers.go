package riskadvisorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/Prytu/risk-advisor/pkg/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func NewAdviseHandler(b *brain.Brain, schedulerCommunicationPort string) HTTPHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		//clusterMutations, err := parseAdviseRequestBody(r.Body)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusBadRequest)
		//	return
		//}

		log.Printf("Starting scheduler server on port %s\n", schedulerCommunicationPort)
		schedHandler := schedulerHandler.New(b, schedulerCommunicationPort)
		go http.ListenAndServe(":"+schedulerCommunicationPort, schedHandler)
		//s := simulator.New(b, schedulerHandler)

		//response := s.RunMultiplePodSimulation(clusterMutations.ToCreate, clusterMutations.ToDelete)
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

func respond(status, message string, w http.ResponseWriter) {
	response := model.ProxyResponse{
		Status:  status,
		Message: message,
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Error marshalling response: %v\n", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
