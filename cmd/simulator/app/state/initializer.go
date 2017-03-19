package state

import (
	"fmt"
	"strconv"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/fields"
	"k8s.io/client-go/1.5/rest"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state/fieldselectors"
)

func InitState() *ClusterState {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintf("Error getting Kubernetes in-cluster config %v", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Error creating Kubernetes client %v", err))
	}

	pvcs, _ := clientset.Core().PersistentVolumeClaims("default").List(api.ListOptions{
		ResourceVersion: "0",
	})

	pvs, _ := clientset.Core().PersistentVolumes().List(api.ListOptions{
		ResourceVersion: "0",
	})

	replicasets, _ := clientset.ExtensionsClient.ReplicaSets("default").List(api.ListOptions{
		ResourceVersion: "0",
	})

	services, _ := clientset.Core().Services("default").List(api.ListOptions{
		ResourceVersion: "0",
	})

	replicationControllers, _ := clientset.Core().ReplicationControllers("default").List(api.ListOptions{
		ResourceVersion: "0",
	})

	var assignedSelector fields.Selector
	stringAssignedSelector := fieldselectors.AssignedNonTerminatedPods
	api.Convert_string_To_fields_Selector(&stringAssignedSelector, &assignedSelector, nil)

	assignedPods, _ := clientset.Core().Pods("default").List(api.ListOptions{
		FieldSelector:   assignedSelector,
		ResourceVersion: "0",
	})

	var unassignedSelector fields.Selector
	stringUnassignedSelector := fieldselectors.UnassignedNonTerminatedPods
	api.Convert_string_To_fields_Selector(&stringUnassignedSelector, &unassignedSelector, nil)

	unassignedPods, _ := clientset.Core().Pods("default").List(api.ListOptions{
		FieldSelector:   unassignedSelector,
		ResourceVersion: "0",
	})

	podMap := make(map[string]v1.Pod, len(assignedPods.Items)+len(unassignedPods.Items))
	for _, pod := range assignedPods.Items {
		podMap[pod.Name] = pod
	}
	for _, pod := range unassignedPods.Items {
		podMap[pod.Name] = pod
	}

	nodeList, _ := clientset.Core().Nodes().List(api.ListOptions{
		ResourceVersion: "0",
	})

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
