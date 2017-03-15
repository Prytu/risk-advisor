package kubeClient

import (
	"errors"
	"net/http"
	"time"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
)

type ClusterCommunicator interface {
	CreatePod(pod *v1.Pod, podName, namespace string) (string, error)
	WaitUntilPodReady(url string) error
	DeletePod(namespace, podName string) error
}

type kubernetesClient struct {
	clientset   *kubernetes.Clientset
	waitTimeout int
}

func New(waitTimeout int) (ClusterCommunicator, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		clientset:   clientset,
		waitTimeout: waitTimeout,
	}, nil
}

func (kc *kubernetesClient) CreatePod(pod *v1.Pod, podName, namespace string) (string, error) {
	_, err := kc.clientset.Core().Pods(namespace).Create(pod)
	if err != nil {
		return "", err
	}

	newPod, err := kc.clientset.Core().Pods(namespace).Get(podName)
	for newPod.Status.PodIP == "" {
		time.Sleep(time.Second)
		newPod, err = kc.clientset.Core().Pods(namespace).Get(podName)
	}

	return newPod.Status.PodIP, nil
}

func (kc *kubernetesClient) WaitUntilPodReady(url string) error {
	_, err := http.Get(url)

	attempts := 1
	for err != nil {
		time.Sleep(time.Second)
		_, err = http.Get(url)

		if attempts == kc.waitTimeout {
			return errors.New("Timed out when waiting for pod to start running.")
		}
		attempts++
	}

	return nil
}

func (kc *kubernetesClient) DeletePod(namespace, podName string) error {
	return kc.clientset.Core().Pods(namespace).Delete(podName, &api.DeleteOptions{})
}
