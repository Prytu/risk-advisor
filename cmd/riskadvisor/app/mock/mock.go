package mock

import (
	"net/http"

	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

// Http Client mocking helpers
type roundTripperMock func(*http.Request) (*http.Response, error)

func (fun roundTripperMock) RoundTrip(r *http.Request) (*http.Response, error) {
	return fun(r)
}

func MockHTTPClient(roundTripper func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: roundTripperMock(roundTripper),
	}
}

// Kubernetes client mock
type KubernetesClientMock struct {
	mock.Mock
}

func (kcm *KubernetesClientMock) CreatePod(pod *v1.Pod, podName, namespace string, timeout int) (string, error) {
	args := kcm.Called(podName)
	return args.String(0), args.Error(1)
}

func (kcm *KubernetesClientMock) WaitUntilPodReady(url string, timeout int) error {
	args := kcm.Called(url)
	return args.Error(0)
}

func (kcm *KubernetesClientMock) DeletePod(namespace, podName string) error {
	args := kcm.Called(podName)
	return args.Error(0)
}
