package podprovider

import (
	"errors"

	"k8s.io/kubernetes/pkg/api"
	"log"
)

// TODO: add synchronization, a queue of pods instead of a single one
type UnscheduledPodProvider interface {
	AddPod(pod *api.Pod) error
	GetPod() (*api.Pod, error)
	Reset() error
}

var NoPods = errors.New("No pods to schedule.")

type SinglePodProvider struct {
	currentPod *api.Pod
}

func New() *SinglePodProvider {
	return &SinglePodProvider{}
}

func (provider *SinglePodProvider) AddPod(pod *api.Pod) error {
	pod.Namespace = "default" // TODO: Maybe get from pod and if empty then assign default
	pod.SelfLink = "/api/v1/namespaces/" + pod.Namespace + "/pods/" + pod.Name

	provider.currentPod = pod

	return nil
}

func (provider *SinglePodProvider) GetPod() (*api.Pod, error) {
	if provider.currentPod == nil {
		return nil, NoPods
	}

	log.Print("GET POD() CALLED")

	pod := provider.currentPod
	provider.currentPod = nil

	return pod, nil
}

func (provider *SinglePodProvider) Reset() error {
	provider.currentPod = nil

	return nil
}
