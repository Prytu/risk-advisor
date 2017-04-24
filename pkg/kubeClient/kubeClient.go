package kubeClient

import (
	"errors"
	"net/http"
	"time"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
	v1beta1 "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/fields"
)

type ClusterCommunicator interface {
	PodOperationHandler
	ClusterStateFetcher
}

type PodOperationHandler interface {
	CreatePod(pod *v1.Pod, podName, namespace string, timeout int) (string, error)
	WaitUntilPodReady(url string, timeout int) error
	DeletePod(namespace, podName string) error
}

type ClusterStateFetcher interface {
	GetPVCs(namespace string) (*v1.PersistentVolumeClaimList, error)
	GetPVs() (*v1.PersistentVolumeList, error)
	GetReplicaSets(namespace string) (*v1beta1.ReplicaSetList, error)
	GetServices(namespace string) (*v1.ServiceList, error)
	GetReplicationControllers(namespace string) (*v1.ReplicationControllerList, error)
	GetPods(namespace string, fieldSelector fields.Selector) (*v1.PodList, error)
	GetNodes() (*v1.NodeList, error)
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
	for err != nil { // Loop here in case of simulator being still in 'Terminating' state after previous request
		time.Sleep(time.Second)
		_, err = kc.clientset.Core().Pods(namespace).Create(pod)

		if time.Now().After(deadline) {
			return "", err
		}
	}

	now = time.Now()
	deadline = now.Add(time.Duration(timeout) * time.Second)

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

func (kc *kubernetesClient) GetPVCs(namespace string) (*v1.PersistentVolumeClaimList, error) {
	return kc.clientset.Core().PersistentVolumeClaims(namespace).List(api.ListOptions{
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetPVs() (*v1.PersistentVolumeList, error) {
	return kc.clientset.Core().PersistentVolumes().List(api.ListOptions{
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetReplicaSets(namespace string) (*v1beta1.ReplicaSetList, error) {
	return kc.clientset.ExtensionsClient.ReplicaSets("default").List(api.ListOptions{
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetServices(namespace string) (*v1.ServiceList, error) {
	return kc.clientset.Core().Services(namespace).List(api.ListOptions{
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetReplicationControllers(namespace string) (*v1.ReplicationControllerList, error) {
	return kc.clientset.Core().ReplicationControllers("default").List(api.ListOptions{
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetPods(namespace string, fieldSelector fields.Selector) (*v1.PodList, error) {
	return kc.clientset.Core().Pods("default").List(api.ListOptions{
		FieldSelector:   fieldSelector,
		ResourceVersion: "0",
	})
}

func (kc *kubernetesClient) GetNodes() (*v1.NodeList, error) {
	return kc.clientset.Core().Nodes().List(api.ListOptions{
		ResourceVersion: "0",
	})
}
