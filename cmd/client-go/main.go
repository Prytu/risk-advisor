package main

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"net/http"
	"io/ioutil"
)

func main() {
	var pod v1.Pod

	pod = v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "simulator",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{v1.Container{
				Name: "simulator",
				Image: "pposkrobko/simulator",
				Ports: []v1.ContainerPort{v1.ContainerPort{ContainerPort: 9997},},
			},},
		},
	}


	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}


	_, err = clientset.Core().Pods("default").Create(&pod)
	if err != nil {
		fmt.Printf(err.Error())
	}


	for {
		time.Sleep(10 * time.Second)

		pods, err := clientset.Core().Pods("").List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		newPod, err := clientset.Core().Pods("default").Get("risk-adv", metav1.GetOptions{})
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		var podIp = newPod.Status.PodIP


		fmt.Printf(podIp)

		resp, err := http.Get("http://" + podIp + ":9997/advise")
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		fmt.Println(string(body))
	}
}
