package riskadvisorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Prytu/risk-advisor/cmd/proxy/app/podprovider"
	"github.com/Prytu/risk-advisor/pkg/model"
	"k8s.io/kubernetes/pkg/api/v1"
	"log"
	"strings"
)

type RiskAdvisorHandler struct {
	server               *http.ServeMux
	proxyResponseChannel <-chan interface{}
	podProvider          podprovider.UnscheduledPodProvider
}

func New(proxyResponseChannel <-chan interface{}, podProvider podprovider.UnscheduledPodProvider) *RiskAdvisorHandler {

	mux := http.NewServeMux()

	adviseHandler := newAdviseHandler(proxyResponseChannel, podProvider)
	mux.HandleFunc("/advise", adviseHandler)

	return &RiskAdvisorHandler{
		server:               mux,
		proxyResponseChannel: proxyResponseChannel,
		podProvider:          podProvider,
	}
}

func (handler *RiskAdvisorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.server.ServeHTTP(w, r)
}

func newAdviseHandler(responseChannel <-chan interface{},
	podProvider podprovider.UnscheduledPodProvider) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		pod, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("\n\nPROXY POD: %v\n", pod)

		if err = podProvider.AddPod(pod); err != nil {
			errorMessage := fmt.Sprintf("Error adding pod: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}

		proxyResponse := <-responseChannel
		switch proxyResponse := proxyResponse.(type) {
		case v1.Binding:
			message := fmt.Sprintf("Pod %s has been sucessfully scheduled on node %v.", pod.Name, proxyResponse.Target.Name)
			respond("Success", message, w)

		case model.FailedSchedulingResponse:
			message := strings.Replace(proxyResponse.Message, "\n", " ", -1)
			respond("Failure", message, w)

		case error:
			log.Print(fmt.Sprintf("Proxy error: %v\n", proxyResponse))
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
		}

		return
	}
}

func parseAdviseRequestBody(requestBody io.ReadCloser) (*v1.Pod, error) {
	body, err := ioutil.ReadAll(requestBody)
	if err != nil {
		errorMessage := fmt.Sprintf("Error reading request body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	var adviceRequest model.AdviceRequest

	err = json.Unmarshal(body, &adviceRequest)
	if err != nil {
		errorMessage := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	return adviceRequest.Pod, nil
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
