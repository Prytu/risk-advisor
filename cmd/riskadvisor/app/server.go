package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/Prytu/risk-advisor/pkg/kubeClient"
	"github.com/Prytu/risk-advisor/pkg/model"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/gorilla/mux.v1"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

type AdviceService struct {
	server                  *mux.Router
	simulatorPort           string
	clusterCommunicator     kubeClient.ClusterCommunicator
	httpClient              http.Client
	simulatorStartupTimeout int
	handlerLock             sync.Mutex
}

func New(simulatorPort string, clusterCommunicator kubeClient.ClusterCommunicator, httpClient http.Client,
	simulatorStartupTimeout int) *AdviceService {
	as := AdviceService{
		server:                  mux.NewRouter(),
		simulatorPort:           simulatorPort,
		clusterCommunicator:     clusterCommunicator,
		httpClient:              httpClient,
		simulatorStartupTimeout: simulatorStartupTimeout,
		handlerLock:             sync.Mutex{},
	}

	as.register()

	return &as
}

func (as *AdviceService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	as.server.ServeHTTP(w, r)
}

func (as *AdviceService) register() {
	as.server.HandleFunc("/advise", as.sendAdviceRequest).Methods("POST")
}

func (as *AdviceService) sendAdviceRequest(w http.ResponseWriter, r *http.Request) {
	as.handlerLock.Lock()
	defer as.handlerLock.Unlock()

	simulatorIP, err := as.startSimulatorPod()
	defer as.cleanup()
	if err != nil {
		writeError(w, fmt.Sprintf("Error starting simulator pod: %s", err))
		return
	}

	log.Print("Sending simulator request")
	simulatorResponse, err := as.sendSimulatorRequest(simulatorIP, r)
	if err != nil {
		writeError(w, fmt.Sprintf("Error communicating with simulator: %s", err))
		return
	}

	log.Print("Received response from simulator")
	riskAdvisorResponse, err := json.MarshalIndent(simulatorResponse, "", " ")
	if err != nil {
		log.WithError(err).Error("Error writing simulator response")
		writeError(w, fmt.Sprint("Unexpected server error."))
		return
	}

	writeStatusCodeAndContentType(w, http.StatusOK)
	w.Write(riskAdvisorResponse)
}

func (as *AdviceService) startSimulatorPod() (string, error) {
	log.Print("Creating simulator pod")
	podIP, err := as.clusterCommunicator.CreatePod(simulatorPod, "simulator", "default", as.simulatorStartupTimeout)
	if err != nil {
		log.WithError(err).Error("error creating simulator pod")
		return "", err
	}

	log.Print("Waiting until simulator is ready")
	err = as.clusterCommunicator.WaitUntilPodReady(as.getSimulatorAdviseUrl(podIP), as.simulatorStartupTimeout)
	if err != nil {
		log.WithError(err).Error("error waiting for simulator pod")
		return "", err
	}

	return podIP, nil
}

func (as *AdviceService) sendSimulatorRequest(podIP string, request *http.Request) ([]model.SchedulingResult, error) {
	simulatorRequestJSON, err := as.generateSimulatorRequest(request)
	if err != nil {
		return nil, err
	}

	resp, err := as.httpClient.Post(
		as.getSimulatorAdviseUrl(podIP),
		"application/json",
		bytes.NewReader(simulatorRequestJSON),
	)
	if err != nil {
		errorMessage := "error performing Post request to simulator"
		log.WithError(err).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	responseJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorMessage := "error reading simulator request"
		log.WithError(err).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var simulatorResponse []model.SchedulingResult
	err = json.Unmarshal(responseJSON, &simulatorResponse)
	if err != nil {
		errorMessage := "error unmarshalling simulator request"
		log.WithError(err).Error(errorMessage)
		return nil, err
	}

	return simulatorResponse, nil
}

func (as *AdviceService) cleanup() {
	log.Print("Deleting simulator pod")

	err := as.clusterCommunicator.DeletePod("default", "simulator")
	if err != nil {
		log.WithError(err).Error("error deleting simulator pod")
	}
}

func (as *AdviceService) generateSimulatorRequest(request *http.Request) ([]byte, error) {
	pods, err := as.getPodsFromRequest(request)
	if err != nil {
		return nil, err
	}

	simulatorRequest := model.SimulatorRequest{ToCreate: pods}
	simulatorRequestJSON, err := json.Marshal(simulatorRequest)
	if err != nil {
		errorMessage := "error marshalling simulatorRequest"
		log.WithError(err).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}
	return simulatorRequestJSON, nil
}

func (as *AdviceService) getPodsFromRequest(request *http.Request) ([]*v1.Pod, error) {
	var pods []*v1.Pod

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		errorMessage := "error reading request body"
		log.WithError(err).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	err = json.Unmarshal(body, &pods)
	if err != nil {
		errorMessage := "error unmarshalling request body"
		log.WithError(err).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	return pods, nil
}

func (as *AdviceService) getSimulatorAdviseUrl(podIP string) string {
	return fmt.Sprintf("http://%s:%s/advise", podIP, as.simulatorPort)
}

func writeError(w http.ResponseWriter, errorMsg string) {
	writeStatusCodeAndContentType(w, http.StatusInternalServerError)
	riskAdvisorResponse, err := json.Marshal(model.SchedulingError{
		errorMsg,
	})
	if err != nil {
		log.WithError(err).Fatal("error while marshalling error message")
	}

	w.Write(riskAdvisorResponse)
}

func writeStatusCodeAndContentType(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
}
