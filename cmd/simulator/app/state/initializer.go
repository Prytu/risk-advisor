package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/kubernetes/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state/urls"
)

// TODO: ugly global, fix
var baseURL string

func InitState(apiserverURL string) *State {
	baseURL = apiserverURL

	var assignedPods v1.PodList
	getResource(urls.AssignedNonTerminatedPods, &assignedPods)
	var unassignedPods v1.PodList
	getResource(urls.UnassignedNonTerminatedPods, &unassignedPods)
	var nodeList v1.NodeList
	getResource(urls.Nodes, &nodeList)

	pvcs := getJSONResource(urls.Pvcs)
	pvs := getJSONResource(urls.Pvs)
	replicasets := getJSONResource(urls.Replicasets)
	services := getJSONResource(urls.Services)
	replicationControllers := getJSONResource(urls.ReplicationControllers)

	return &State{
		Pods:                   append(assignedPods.Items, unassignedPods.Items...),
		Nodes:                  nodeList.Items,
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
	req, err := http.NewRequest("GET", baseURL+url, nil)
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
