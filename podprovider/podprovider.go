package podprovider

import (
	"errors"

	"k8s.io/kubernetes/pkg/api"
	"log"
	"strconv"
	"sync"
)

// TODO: add synchronization, a queue of pods instead of a single pod?
type UnscheduledPodProvider interface {
	AddPod(pod *api.Pod) error
	GetPod() (api.Pod, error)
	Reset() error
}

var NoPods = errors.New("No pods to schedule.")

type SinglePodProvider struct {
	resourceVersion int64
	currentPod      *api.Pod
	mutex           *sync.Mutex
}

func New() *SinglePodProvider {
	return &SinglePodProvider{
		resourceVersion: 0,
		mutex:           &sync.Mutex{},
	}
}

func (provider *SinglePodProvider) AddPod(pod *api.Pod) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.resourceVersion += 1

	pod.Namespace = "default" // TODO: get from pod and if empty then assign default
	pod.SelfLink = "/api/v1/namespaces/" + pod.Namespace + "/pods/" + pod.Name
	pod.ResourceVersion = strconv.FormatInt(provider.resourceVersion, 10)
	pod.Status.Phase = "Pending"

	provider.currentPod = pod

	return nil
}

func (provider *SinglePodProvider) GetPod() (api.Pod, error) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if provider.currentPod == nil {
		return api.Pod{}, NoPods
	}

	pod := *provider.currentPod
	provider.currentPod = nil

	return pod, nil
}

func (provider *SinglePodProvider) Reset() error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.currentPod = nil

	return nil
}
