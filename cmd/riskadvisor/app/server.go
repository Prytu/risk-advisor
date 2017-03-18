package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Prytu/risk-advisor/pkg/kubeClient"
	"github.com/Prytu/risk-advisor/pkg/model"

	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

type AdviceService struct {
	simulatorPort           string
	clusterCommunicator     kubeClient.ClusterCommunicator
	httpClient              http.Client
	simulatorStartupTimeout int
}

func New(simulatorPort string, clusterCommunicator kubeClient.ClusterCommunicator, httpClient http.Client,
	simulatorStartupTimeout int) http.Handler {
	as := AdviceService{
		simulatorPort:           simulatorPort,
		clusterCommunicator:     clusterCommunicator,
		httpClient:              httpClient,
		simulatorStartupTimeout: simulatorStartupTimeout,
	}

	wsContainer := restful.NewContainer()
	as.Register(wsContainer)

	return wsContainer
}

func (as *AdviceService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/advise").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("").To(as.sendAdviceRequest).
		Doc("Post a request for advice").
		Reads([]v1.Pod{}).
		Returns(http.StatusOK, "OK", []model.SchedulingResult{}))

	container.Add(ws)
}

func (as *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {
	simulatorIP, err := as.startSimulatorPod()
	defer as.cleanup()
	if err != nil {
		response.WriteHeaderAndEntity(
			http.StatusInternalServerError,
			model.SchedulingError{
				fmt.Sprintf("Error starting simulator pod: %s", err),
			},
		)
		return
	}

	log.Print("Sending simulator request")
	simulatorResponse, err := as.sendSimulatorRequest(simulatorIP, request)
	if err != nil {
		response.WriteHeaderAndEntity(
			http.StatusInternalServerError,
			model.SchedulingError{
				fmt.Sprintf("Error communicating with simulator: %s", err),
			},
		)
		return
	}

	log.Print("Received response from simulator")
	err = response.WriteEntity(simulatorResponse)
	if err != nil {
		log.WithError(err).Error("error writing response")
		response.WriteHeaderAndEntity(
			http.StatusInternalServerError,
			model.SchedulingError{
				fmt.Sprint("Unexpected server error."),
			},
		)
		return
	}
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

func (as *AdviceService) sendSimulatorRequest(podIP string, request *restful.Request) ([]model.SchedulingResult, error) {
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

func (as *AdviceService) generateSimulatorRequest(request *restful.Request) ([]byte, error) {
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

func (as *AdviceService) getPodsFromRequest(request *restful.Request) ([]*v1.Pod, error) {
	var pods []*v1.Pod

	body, err := ioutil.ReadAll(request.Request.Body)
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
