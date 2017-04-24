package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mocks "github.com/Prytu/risk-advisor/cmd/riskadvisor/app/mock"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
	"github.com/Prytu/risk-advisor/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"k8s.io/client-go/1.5/pkg/api/v1"
)

func TestSuccess(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser([]*v1.Pod{}))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("podIP", nil).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(nil).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	expectedBody := []model.SchedulingResult{{PodName: "pod", Result: "success", Message: "success"}}
	simulatorResponse := createHTTPClientSuccessResponseFunc(http.StatusOK, expectedBody, defaultHeader())
	adviceService := createServiceWithMockHttpClient(simulatorResponse, clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	expectedBodyBytes, err := json.MarshalIndent(expectedBody, "", " ")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Equal(t, recorder.Body.Bytes(), expectedBodyBytes)
}

func TestCommunicationWithSimulatorFailure(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser([]*v1.Pod{}))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("podIP", nil).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(nil).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	simulatorResponse := createHTTPClientErrorResponseFunc(communicationWithSimulatorErrorMessage)
	adviceService := createServiceWithMockHttpClient(simulatorResponse, clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	expectedBody := model.SchedulingResult{
		ErrorMessage: fmt.Sprintf("Error communicating with simulator: %s", communicationWithSimulatorErrorMessage),
	}
	expectedBodyBytes, err := json.Marshal(expectedBody)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Equal(t, recorder.Body.Bytes(), expectedBodyBytes)
}

func TestCreatingPodFailure(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser([]*v1.Pod{}))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New(creatingPodErrorMessage)).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(nil).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	adviceService := createService(clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	expectedBody := model.SchedulingResult{
		ErrorMessage: fmt.Sprintf("Error starting simulator pod: %s", creatingPodErrorMessage),
	}
	expectedBodyBytes, err := json.Marshal(expectedBody)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Equal(t, recorder.Body.Bytes(), expectedBodyBytes)
}

func TestWaitingUntilPodReadyFailure(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser([]*v1.Pod{}))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("podIP", nil).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(errors.New(waitingUntilPodReadyErrorMessage)).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	adviceService := createService(clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	expectedBody := model.SchedulingResult{
		ErrorMessage: fmt.Sprintf("Error starting simulator pod: %s", waitingUntilPodReadyErrorMessage),
	}
	expectedBodyBytes, err := json.Marshal(expectedBody)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Equal(t, recorder.Body.Bytes(), expectedBodyBytes)
}

func TestUsersIncorrectRequest(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser(
		ioutil.NopCloser(bytes.NewBuffer([]byte("some not unmarshallable thing")))))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("podIP", nil).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(nil).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	ignoredSimulatorResponseBody := []model.SchedulingResult{{PodName: "pod", Result: "success", Message: "success"}}
	ignoredSimulatorResponse := createHTTPClientSuccessResponseFunc(http.StatusOK, ignoredSimulatorResponseBody, defaultHeader())
	adviceService := createServiceWithMockHttpClient(ignoredSimulatorResponse, clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	expectedBody := model.SchedulingResult{
		ErrorMessage: fmt.Sprintf("Error communicating with simulator: %s", unmarshallingRequestBodyErrorMessage),
	}
	expectedBodyBytes, err := json.Marshal(expectedBody)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Equal(t, recorder.Body.Bytes(), expectedBodyBytes)
}

func TestIncorrectSimulatorsResponse(t *testing.T) {
	request, _ := http.NewRequest("POST", "/advise", bodyToReadCloser([]*v1.Pod{}))

	clusterCommunicatorMock := &mocks.KubernetesClientMock{}
	clusterCommunicatorMock.
		On("CreatePod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("podIP", nil).
		On("WaitUntilPodReady", mock.Anything, mock.Anything).Return(nil).
		On("DeletePod", mock.Anything, mock.Anything).Return(nil)
	simulatorResponse := createHTTPClientIncorrectResponseFunc()
	adviceService := createServiceWithMockHttpClient(simulatorResponse, clusterCommunicatorMock)

	recorder := httptest.NewRecorder()
	adviceService.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Header()["Content-Type"], "application/json")
	assert.Contains(t, string(recorder.Body.Bytes()), "Error communicating with simulator")
}

func createService(
	clusterCommunicatorMock kubeClient.PodOperationHandler,
) *AdviceService {
	return New(defaults.SimulatorPort, clusterCommunicatorMock, http.Client{}, defaults.StartupTimeout)
}

type HttpClientResponseFunc func(*http.Request) (*http.Response, error)

func createServiceWithMockHttpClient(
	simulatorResponseMockFunc HttpClientResponseFunc,
	clusterCommunicatorMock kubeClient.PodOperationHandler,
) *AdviceService {
	httpClient := mocks.MockHTTPClient(simulatorResponseMockFunc)
	return New(defaults.SimulatorPort, clusterCommunicatorMock, *httpClient, defaults.StartupTimeout)
}

func createHTTPClientSuccessResponseFunc(
	statusCode int,
	body []model.SchedulingResult,
	header http.Header,
) HttpClientResponseFunc {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Body:       bodyToReadCloser(body),
			Header:     header,
		}, nil
	}
}

func createHTTPClientIncorrectResponseFunc() HttpClientResponseFunc {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("some not unmarshallable thing"))),
			Header:     defaultHeader(),
		}, nil
	}
}

func createHTTPClientErrorResponseFunc(errorMessage string) HttpClientResponseFunc {
	return func(r *http.Request) (*http.Response, error) {
		return nil, errors.New(errorMessage)
	}
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	return header
}

func bodyToReadCloser(bodyData interface{}) io.ReadCloser {
	data, err := json.Marshal(bodyData)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(data)

	return ioutil.NopCloser(buf)
}

const creatingPodErrorMessage = "pod could not be created"
const waitingUntilPodReadyErrorMessage = "error while waiting until pod ready"
const communicationWithSimulatorErrorMessage = "error performing Post request to simulator"
const unmarshallingRequestBodyErrorMessage = "error unmarshalling request body"
