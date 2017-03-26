package riskadvisorhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/schedulerHandler"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/simulator"
	"github.com/Prytu/risk-advisor/pkg/model"
)

type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func MultiplePodAdviseHandler(b *brain.Brain, schedulerCommunicationPort string, eventChannel chan *v1.Event) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterMutations, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			errorMsg := "invalid request body"
			log.WithError(err).Error(errorMsg)
			respondWithError(w, fmt.Sprintf("%s (%s)", errorMsg, err), http.StatusBadRequest)
			return
		}

		schedHandler := schedulerHandler.New(b, schedulerCommunicationPort)
		s := simulator.New(b, schedHandler, eventChannel)

		result := s.RunMultiplePodSimulation(clusterMutations.ToCreate, clusterMutations.ToDelete)

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			errorMsg := "error marshalling response"
			log.WithError(err).Error(errorMsg)
			respondWithError(w, fmt.Sprintf("%s (%s)", errorMsg, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(resultJSON)
	}
}

func ErrorResponseHandler(err error) HTTPHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorMessage := fmt.Sprintf("Error during simulator initalization: %s.", err)
		respondWithError(w, errorMessage, http.StatusInternalServerError)
	}
}

func parseAdviseRequestBody(requestBody io.ReadCloser) (*model.SimulatorRequest, error) {
	body, err := ioutil.ReadAll(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %s", err)
	}

	var adviceRequest model.SimulatorRequest

	err = json.Unmarshal(body, &adviceRequest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling body: %s", err)
	}

	return &adviceRequest, nil
}

func respondWithError(w http.ResponseWriter, appError string, statusCode int) {
	errStruct := model.SchedulingError{
		ErrorMessage: appError,
	}

	errJSON, err := json.MarshalIndent(errStruct, "", "  ")
	if err != nil {
		errorMsg := fmt.Sprintf("Error marshalling error response: %s of application error: %s.", err, appError)
		log.WithError(err).Error(errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError) // Just answer with text/plain message

		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(errJSON)
}
