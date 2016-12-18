package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"k8s.io/kubernetes/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state/urls"
)

func InitState(apiserverURL string) *ClusterState {
	pvcs := getJSONResource(apiserverURL + urls.Pvcs)
	pvs := getJSONResource(apiserverURL + urls.Pvs)
	replicasets := getJSONResource(apiserverURL + urls.Replicasets)
	services := getJSONResource(apiserverURL + urls.Services)
	replicationControllers := getJSONResource(apiserverURL + urls.ReplicationControllers)

	var assignedPods v1.PodList
	var unassignedPods v1.PodList
	getResource(apiserverURL+urls.AssignedNonTerminatedPods, &assignedPods)
	getResource(apiserverURL+urls.UnassignedNonTerminatedPods, &unassignedPods)

	podMap := make(map[string]v1.Pod, len(assignedPods.Items)+len(unassignedPods.Items))
	for _, pod := range assignedPods.Items {
		podMap[pod.Name] = pod
	}
	for _, pod := range unassignedPods.Items {
		podMap[pod.Name] = pod
	}

	var nodeList v1.NodeList
	getResource(apiserverURL+urls.Nodes, &nodeList)

	nodeMap := make(map[string]v1.Node, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodeMap[node.Name] = node
	}

	resourceVersion, err := strconv.ParseInt(nodeList.ResourceVersion, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Error parsing resourceVersion: %v", err))
	}

	return &ClusterState{
		resourceVersion:        resourceVersion,
		pods:                   podMap,
		nodes:                  nodeMap,
		Pvcs:                   pvcs,
		Pvs:                    pvs,
		Replicasets:            replicasets,
		Services:               services,
		ReplicationControllers: replicationControllers,
	}
}

func getResource(url string, resource interface{}) interface{} {
	respJSON := getJSONResource(url)
	err := json.Unmarshal(respJSON, resource)
	if err != nil {
		panic(fmt.Sprintf("get %v resp unmarshall error: %v", url, err))
	}

	return resource
}

func getJSONResource(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(fmt.Sprintf("create request %v error: %v", url, err))
	}
	req.Header.Add("Content-Type", `application/json`)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("get %v error: %v", url, err))
	}

	respJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("get %v resp ReadAll error: %v", url, err))
	}

	return respJSON
}
