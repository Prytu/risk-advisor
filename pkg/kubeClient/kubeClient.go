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
	CreatePod(pod *v1.Pod, podName, namespace string, timeout int) (string, error)
	WaitUntilPodReady(url string, timeout int) error
	DeletePod(namespace, podName string) error
}

type kubernetesClient struct {
	clientset  *kubernetes.Clientset
	httpClient http.Client
}

func New(httpClient http.Client) (ClusterCommunicator, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesClient{
		clientset:  clientset,
		httpClient: httpClient,
	}, nil
}

func (kc *kubernetesClient) CreatePod(pod *v1.Pod, podName, namespace string, timeout int) (string, error) {
	now := time.Now()
	deadline := now.Add(time.Duration(timeout) * time.Second)

	_, err := kc.clientset.Core().Pods(namespace).Create(pod)
	if err != nil {
		return "", err
	}

	newPod, err := kc.clientset.Core().Pods(namespace).Get(podName)
	for newPod.Status.PodIP == "" {
		time.Sleep(time.Second)
		newPod, err = kc.clientset.Core().Pods(namespace).Get(podName)

		if time.Now().After(deadline) {
			return "", errors.New("timed out when waiting for simulator pod to be assigned an IP")
		}
	}

	return newPod.Status.PodIP, nil
}

func (kc *kubernetesClient) WaitUntilPodReady(url string, timeout int) error {
	now := time.Now()
	deadline := now.Add(time.Duration(timeout) * time.Second)

	_, err := kc.httpClient.Get(url)

	for err != nil {
		time.Sleep(time.Second)
		_, err = kc.httpClient.Get(url)

		if time.Now().After(deadline) {
			return errors.New("timed out when waiting for simulator pod to start running")
		}
	}

	return nil
}

func (kc *kubernetesClient) DeletePod(namespace, podName string) error {
	return kc.clientset.Core().Pods(namespace).Delete(podName, &api.DeleteOptions{})
}
