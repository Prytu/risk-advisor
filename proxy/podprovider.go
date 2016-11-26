package proxy

import (
	"encoding/json"
	"fmt"
	"k8s.io/kubernetes/pkg/api"
)

type UnscheduledPodProvider interface {
	ProvidePods() *api.PodList
}

type FakePodProvider struct{}

func (provider *FakePodProvider) ProvidePods() *api.PodList {
	podListWithPod := "{\"kind\":\"PodList\",\"apiVersion\":\"v1\",\"metadata\":{\"selfLink\":\"/api/v1/pods\"," +
		"\"resourceVersion\":\"617\"},\"items\":[" +
		"{\"metadata\":{\"name\":\"nginx-without-nodename\",\"namespace\":\"default\"," +
		"\"selfLink\":\"/api/v1/namespaces/default/pods/nginx-without-nodename\"," +
		"\"uid\":\"b0220242-b346-11e6-a633-000c2999b232\",\"resourceVersion\":\"614\"," +
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

	err := json.Unmarshal([]byte(responses[1]), &podList)
	if err != nil {
		panic(fmt.Sprintf("error marshalling pod: %v\n", err))
	}

	return &podList
}
