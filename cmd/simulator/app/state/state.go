package state

import (
	"fmt"
	
	"k8s.io/kubernetes/pkg/api/v1"
)

type State struct {
	Pods                   []v1.Pod
	Nodes                  []v1.Node
	Pvcs                   []byte
	Pvs                    []byte
	Replicasets            []byte
	Services               []byte
	ReplicationControllers []byte
}

func (s *State) String() string {
	return fmt.Sprintf(`State:
		Pods:%v
		Nodes:%v
		Pvcs:%v
		Pvs:%v
		Replicasets:%v
		Services:%v
		ReplcationControllers:%v`,
		s.Pods, s.Nodes, s.Pvcs, s.Pvs, s.Replicasets, s.Services, s.ReplicationControllers)
}
