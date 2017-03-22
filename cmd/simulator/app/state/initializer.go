package state

import (
	"fmt"
	"strconv"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state/fieldselectors"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/fields"
)

func InitState(ksf kubeClient.ClusterStateFetcher) (*ClusterState, error) {
	assignedSelector, err := convertFieldSelector(fieldselectors.AssignedNonTerminatedPods)
	if err != nil {
		return nil, err
	}

	unassignedSelector, err := convertFieldSelector(fieldselectors.UnassignedNonTerminatedPods)
	if err != nil {
		return nil, err
	}

	pvcs, err := ksf.GetPVCs("default")
	if err != nil {
		return nil, fmt.Errorf("error fetching PVCs: %s", err)
	}

	pvs, err := ksf.GetPVs()
	if err != nil {
		return nil, fmt.Errorf("error fetching PVs: %s", err)
	}

	replicasets, err := ksf.GetReplicaSets("default")
	if err != nil {
		return nil, fmt.Errorf("error fetching ReplicaSets: %s", err)
	}

	services, err := ksf.GetServices("default")
	if err != nil {
		return nil, fmt.Errorf("error fetching Services: %s", err)
	}

	replicationControllers, err := ksf.GetReplicationControllers("default")
	if err != nil {
		return nil, fmt.Errorf("error fetching Replication controllers: %s", err)
	}

	assignedPods, err := ksf.GetPods("default", assignedSelector)
	if err != nil {
		return nil, fmt.Errorf("error fetching Assigned Pods: %s", err)
	}

	unassignedPods, err := ksf.GetPods("default", unassignedSelector)
	if err != nil {
		return nil, fmt.Errorf("error fetching Unassigned Pods: %s", err)
	}

	nodeList, err := ksf.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("error fetching Nodes Pods: %s", err)
	}

	podMap := make(map[string]v1.Pod, len(assignedPods.Items)+len(unassignedPods.Items))
	for _, pod := range assignedPods.Items {
		podMap[pod.Name] = pod
	}
	for _, pod := range unassignedPods.Items {
		podMap[pod.Name] = pod
	}

	nodeMap := make(map[string]v1.Node, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodeMap[node.Name] = node
	}

	resourceVersion, err := strconv.ParseInt(nodeList.ResourceVersion, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing resourceVersion: %s", err)
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
	}, nil
}

func convertFieldSelector(selectorString string) (fields.Selector, error) {
	var selector fields.Selector
	err := api.Convert_string_To_fields_Selector(&selectorString, &selector, nil)
	if err != nil {
		return nil, fmt.Errorf("error converting FieldSelector %s: %s", selectorString, err)
	}

	return selector, nil
}
