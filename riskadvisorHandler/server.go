package riskadvisorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"k8s.io/kubernetes/pkg/api"

	"github.com/Prytu/risk-advisor/podprovider"
)

type RiskAdvisorHandler struct {
	server               *http.ServeMux
	ProxyResponseChannel chan api.Binding
	PodProvider          podprovider.PodProvider
}

func New(proxyResponseChannel chan api.Binding, podProvider podprovider.PodProvider) *RiskAdvisorHandler {
	mux := http.NewServeMux()
	mux.HandleFunc("/advise", newAdviseHandler(proxyResponseChannel, podProvider))

	return &RiskAdvisorHandler{
		server:               mux,
		ProxyResponseChannel: proxyResponseChannel,
		PodProvider:          podProvider,
	}
}

func (handler *RiskAdvisorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.server.ServeHTTP(w, r)
}

func newAdviseHandler(responseChannel chan api.Binding, podProvider podprovider.PodProvider) func(responseWriter http.ResponseWriter, request *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pod, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			errorMessage := fmt.Sprintf("Error marshalling response: %v\n", err)
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		if err = podProvider.AddPod(pod); err != nil {
			errorMessage := fmt.Sprintf("Error marshalling response: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}

		proxyResponse := <-responseChannel

		json, err := json.MarshalIndent(proxyResponse, "", "  ")
		if err != nil {
			errorMessage := fmt.Sprintf("Error marshalling response: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}

func parseAdviseRequestBody(requestBody io.ReadCloser) (*api.Pod, error) {
	body, err := ioutil.ReadAll(requestBody)
	if err != nil {
		errorMessage := fmt.Sprintf("Error reading request body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	var pod api.Pod

	err = json.Unmarshal(body, &pod)
	if err != nil {
		errorMessage := fmt.Sprintf("Error unmarshalling body: %v\n", err)
		return nil, errors.New(errorMessage)
	}

	return &pod, nil
}
