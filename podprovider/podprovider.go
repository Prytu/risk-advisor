package podprovider

import (
	"errors"

	"k8s.io/kubernetes/pkg/api"
	"log"
	"strconv"
	"sync"
)

// TODO: add synchronization, a queue of pods instead of a single one
type UnscheduledPodProvider interface {
	AddPod(pod *api.Pod) error
	GetPod() (api.Pod, error)
	Reset() error
}

var NoPods = errors.New("No pods to schedule.")

type SinglePodProvider struct {
	resourceVersion int64
	currentPod      *api.Pod
	Mutex           *sync.Mutex
}

func New() *SinglePodProvider {
	return &SinglePodProvider{
		resourceVersion: 0,
		Mutex:           &sync.Mutex{},
	}
}

func (provider *SinglePodProvider) AddPod(pod *api.Pod) error {
	provider.Mutex.Lock()
	defer provider.Mutex.Unlock()

	provider.resourceVersion += 1

	pod.Namespace = "default" // TODO: Maybe get from pod and if empty then assign default
	pod.SelfLink = "/api/v1/namespaces/" + pod.Namespace + "/pods/" + pod.Name
	pod.ResourceVersion = strconv.FormatInt(provider.resourceVersion, 10)
	pod.Status.Phase = "Pending"

	provider.currentPod = pod

	log.Printf("Added POD: %v\n", provider.currentPod)

	return nil
}

func (provider *SinglePodProvider) GetPod() (api.Pod, error) {
	provider.Mutex.Lock()
	defer provider.Mutex.Unlock()

	if provider.currentPod == nil {
		return api.Pod{}, NoPods
	}

	log.Print("GET POD() CALLED")

	pod := *provider.currentPod
	provider.currentPod = nil

	return pod, nil
}

func (provider *SinglePodProvider) Reset() error {
	provider.Mutex.Lock()
	defer provider.Mutex.Unlock()

	provider.currentPod = nil

	return nil
}
