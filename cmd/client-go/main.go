package main

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

func main() {


	podDef := "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"risk-adv\"},\"spec\":{\"containers\":[{\"name\":\"risk-adv\",\"image\":\"pposkrobko/riskadvisor\",\"ports\":[{\"containerPort\":9997}],\"volumeMounts\":[{\"mountPath\":\"/var/log/nginx\",\"name\":\"nginx-logs\"}]},{\"name\":\"log-truncator\",\"image\":\"busybox\",\"command\":[\"/bin/sh\"],\"args\":[\"-c\",\"while true; do cat /dev/null > /logdir/access.log; sleep 10; done\"],\"volumeMounts\":[{\"mountPath\":\"/logdir\",\"name\":\"nginx-logs\"}]}],\"volumes\":[{\"name\":\"nginx-logs\",\"emptyDir\":{}}]}}"


	var pod v1.Pod

	err := json.Unmarshal([]byte(podDef), &pod)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(pod.Name)


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