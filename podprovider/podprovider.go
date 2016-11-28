package podprovider

import (
	"errors"

	"k8s.io/kubernetes/pkg/api"
)

type PodProvider interface {
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
	provider.currentPod = pod

	return nil
}

func (provider *SinglePodProvider) GetPod() (*api.Pod, error) {
	if provider.currentPod == nil {
		return nil, NoPods
	}

	return provider.currentPod, nil
}

func (provider *SinglePodProvider) Reset() error {
	provider.currentPod = nil

	return nil
}
