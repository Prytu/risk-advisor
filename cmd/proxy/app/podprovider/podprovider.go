package podprovider

import (
	"errors"
	"strconv"
	"sync"

	"github.com/Prytu/risk-advisor/cmd/proxy/app/watcher"
	"k8s.io/kubernetes/pkg/api/v1"
)

// TODO: add synchronization, a queue of pods instead of a single pod?
type UnscheduledPodProvider interface {
	AddPod(pod *v1.Pod) error
	GetPod() (v1.Pod, error)
	Reset() error
	SetWatcher(*watcher.Watcher) error
}

var NoPods = errors.New("No pods to schedule.")

type SinglePodProvider struct {
	resourceVersion int64
	currentPod      *v1.Pod
	mutex           *sync.Mutex
	watcher         *watcher.Watcher
}

func New() *SinglePodProvider {
	return &SinglePodProvider{
		resourceVersion: 0,
		mutex:           &sync.Mutex{},
	}
}

func (provider *SinglePodProvider) AddPod(pod *v1.Pod) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.resourceVersion += 1

	pod.Namespace = "default" // TODO: get from pod and if empty then assign default
	pod.SelfLink = "/api/v1/namespaces/" + pod.Namespace + "/pods/" + pod.Name
	pod.ResourceVersion = strconv.FormatInt(provider.resourceVersion, 10)
	pod.Status.Phase = "Pending"

	provider.currentPod = pod
	provider.watcher.Add(pod)

	return nil
}

func (provider *SinglePodProvider) GetPod() (v1.Pod, error) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if provider.currentPod == nil {
		return v1.Pod{}, NoPods
	}

	pod := *provider.currentPod

	return pod, nil
}

func (provider *SinglePodProvider) Reset() error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.watcher.Reset()
	provider.currentPod = nil

	return nil
}

func (provider *SinglePodProvider) SetWatcher(w *watcher.Watcher) error {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.watcher = w

	return nil
}
