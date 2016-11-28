package proxy

import (
	"encoding/json"
	"fmt"
	"k8s.io/kubernetes/pkg/api"
)

type UnscheduledPodProvider interface {
	ProvidePod() *api.Pod
	ProvidePodList() *api.PodList
}

type FakePodProvider struct{}

const podJSON = "{\"metadata\":{\"name\":\"nginx-without-nodename\",\"namespace\":\"default\",\"selfLink\":\"/api/v1/namespaces/default/pods/nginx-without-nodename\",\"uid\":\"0760156d-b4ef-11e6-a051-000c2999b232\",\"resourceVersion\":\"32\",\"creationTimestamp\":\"2016-11-27T22:15:39Z\"},\"spec\":{\"volumes\":[{\"name\":\"nginx-logs\",\"emptyDir\":{}}],\"containers\":[{\"name\":\"nginx\",\"image\":\"nginx\",\"ports\":[{\"containerPort\":80,\"protocol\":\"TCP\"}],\"resources\":{},\"volumeMounts\":[{\"name\":\"nginx-logs\",\"mountPath\":\"/var/log/nginx\"}],\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"Always\"},{\"name\":\"log-truncator\",\"image\":\"busybox\",\"command\":[\"/bin/sh\"],\"args\":[\"-c\",\"while true; do cat /dev/null \\u003e /logdir/access.log; sleep 10; done\"],\"resources\":{},\"volumeMounts\":[{\"name\":\"nginx-logs\",\"mountPath\":\"/logdir\"}],\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"Always\"}],\"restartPolicy\":\"Always\",\"terminationGracePeriodSeconds\":30,\"dnsPolicy\":\"ClusterFirst\",\"securityContext\":{}},\"status\":{\"phase\":\"Pending\"}}"

//const podJSON = "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"nginx-without-nodename\"},\"spec\":{\"containers\":[{\"name\":\"nginx\",\"image\":\"nginx\",\"ports\":[{\"containerPort\":80}],\"volumeMounts\":[{\"mountPath\":\"/var/log/nginx\",\"name\":\"nginx-logs\"}]},{\"name\":\"log-truncator\",\"image\":\"busybox\",\"command\":[\"/bin/sh\"],\"args\":[\"-c\",\"while true; do cat /dev/null > /logdir/access.log; sleep 10; done\"],\"volumeMounts\":[{\"mountPath\":\"/logdir\",\"name\":\"nginx-logs\"}]}],\"volumes\":[{\"name\":\"nginx-logs\",\"emptyDir\":{}}]}}"

var unmarshalledPod *api.Pod

func init() {
	var pod api.Pod

	if err := json.Unmarshal([]byte(podJSON), &pod); err != nil {
		panic(fmt.Sprintf("error marshalling pod: %v\n", err))
	}

	unmarshalledPod = &pod
}

func (provider *FakePodProvider) ProvidePod() *api.Pod {
	return unmarshalledPod
}

func (provider *FakePodProvider) ProvidePodList() *api.PodList {
	podListWithPod := "{\"kind\":\"PodList\",\"apiVersion\":\"v1\",\"metadata\":{\"selfLink\":\"/api/v1/pods\"," +
		"\"resourceVersion\":\"617\"},\"items\":[" +
		"{\"metadata\":{\"name\":\"nginx-without-nodename\",\"namespace\":\"default\"," +
		"\"selfLink\":\"/api/v1/namespaces/default/pods/nginx-without-nodename\"," +
		"\"uid\":\"b0220242-b346-11e6-a633-000c2999b232\",\"resourceVersion\":\"1\"," +
		"\"creationTimestamp\":\"2016-11-25T19:38:06Z\"},\"spec\":{\"volumes\":[{\"name\":\"nginx-logs\"," +
		"\"emptyDir\":{}}],\"containers\":[{\"name\":\"nginx\",\"image\":\"nginx\"," +
		"\"ports\":[{\"containerPort\":80,\"protocol\":\"TCP\"}],\"resources\":{}," +
		"\"volumeMounts\":[{\"name\":\"nginx-logs\",\"mountPath\":\"/var/log/nginx\"}]," +
		"\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"Always\"}," +
		"{\"name\":\"log-truncator\",\"image\":\"busybox\",\"command\":[\"/bin/sh\"]," +
		"\"args\":[\"-c\",\"while true; do cat /dev/null \\u003e /logdir/access.log; sleep 10; done\"]," +
		"\"resources\":{},\"volumeMounts\":[{\"name\":\"nginx-logs\",\"mountPath\":\"/logdir\"}]," +
		"\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"Always\"}]," +
		"\"restartPolicy\":\"Always\",\"terminationGracePeriodSeconds\":30,\"dnsPolicy\":\"ClusterFirst\"," +
		"\"securityContext\":{}},\"status\":{\"phase\":\"Pending\"}}]}"

	emptyPodList := "{" +
		"\"kind\": \"PodList\"," +
		"\"apiVersion\": \"v1\"," +
		"\"metadata\": {" +
		"\"selfLink\": \"/api/v1/pods\"," +
		"\"resourceVersion\": \"30\"}," +
		"\"items\": []}"

	responses := []string{
		podListWithPod,
		emptyPodList,
	}

	var podList api.PodList

	//index := rand.Int() % len(responses)
	//log.Printf("Index = %d, response = %v", index, responses[index])
	//response := responses[index]

	if err := json.Unmarshal([]byte(responses[0]), &podList); err != nil {
		panic(fmt.Sprintf("error marshalling podList: %v\n", err))
	}

	return &podList
}
