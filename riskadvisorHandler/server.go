package riskadvisorhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/api"
	"net/http"

	"github.com/Prytu/risk-advisor/podprovider"
)

type RiskAdvisorHandler struct {
	server               *http.ServeMux
	proxyResponseChannel <-chan api.Binding
	errorChannel         <-chan error
	podProvider          podprovider.UnscheduledPodProvider
}

func New(proxyResponseChannel <-chan api.Binding, errorChannel <-chan error,
	podProvider podprovider.UnscheduledPodProvider) *RiskAdvisorHandler {

	mux := http.NewServeMux()

	adviseHandler := newAdviseHandler(proxyResponseChannel, errorChannel, podProvider)
	mux.HandleFunc("/advise", adviseHandler)

	return &RiskAdvisorHandler{
		server:               mux,
		proxyResponseChannel: proxyResponseChannel,
		errorChannel:         errorChannel,
		podProvider:          podProvider,
	}
}

func (handler *RiskAdvisorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.server.ServeHTTP(w, r)
}

func newAdviseHandler(responseChannel <-chan api.Binding, errorChannel <-chan error,
	podProvider podprovider.UnscheduledPodProvider) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		pod, err := parseAdviseRequestBody(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = podProvider.AddPod(pod); err != nil {
			errorMessage := fmt.Sprintf("Error adding pod: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}

		select {
		case proxyResponse := <-responseChannel:
			responseJSON, err := json.MarshalIndent(proxyResponse, "", "  ")
			if err != nil {
				errorMessage := fmt.Sprintf("Error marshalling response: %v\n", err)
				http.Error(w, errorMessage, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(responseJSON)

		case err := <-errorChannel:
			errorMessage := fmt.Sprintf("Proxy error: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
			return
		}

		return
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
