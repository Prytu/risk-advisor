package state

import (
	"fmt"
	"sync"

	"k8s.io/kubernetes/pkg/api/v1"
)

type ClusterState struct {
	sync.RWMutex
	pods                   map[string]*v1.Pod
	nodes                  map[string]*v1.Node
	pvcs                   []byte
	pvs                    []byte
	replicasets            []byte
	services               []byte
	replicationControllers []byte
}

func (s *ClusterState) AddPods(pods []*v1.Pod) {
	s.Lock()
	defer s.Unlock()

	for _, pod := range pods {
		if _, ok := s.pods[pod.Name]; !ok {
			s.pods[pod.Name] = pod
		}
		// TODO: what to do when the pod with given name is already in the cluster
	}
}

func (s *ClusterState) RemovePods(pods []*v1.Pod) {
	s.Lock()
	defer s.Unlock()

	for _, pod := range pods {
		delete(s.pods, pod.Name)
	}
}

func (s *ClusterState) String() string {
	return fmt.Sprintf(`State:
		Pods:%v
		Nodes:%v
		Pvcs:%v
		Pvs:%v
		Replicasets:%v
		Services:%v
		ReplcationControllers:%v`,
		s.pods, s.nodes, s.pvcs, s.pvs, s.replicasets, s.services, s.replicationControllers)
}
