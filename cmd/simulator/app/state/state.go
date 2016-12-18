package state

import (
	"fmt"
	"sync"

	"k8s.io/kubernetes/pkg/api/v1"
)

type PodFilter func(pod *v1.Pod) bool

type ClusterState struct {
	sync.RWMutex
	resourceVersion int64

	pods                   map[string]v1.Pod
	nodes                  map[string]v1.Node
	Pvcs                   []byte
	Pvs                    []byte
	Replicasets            []byte
	Services               []byte
	ReplicationControllers []byte
}

func (s *ClusterState) AddPod(pod v1.Pod) {
	s.Lock()
	defer s.Unlock()

	s.pods[pod.Name] = pod
}

func (s *ClusterState) GetResourceVersion() int64 {
	s.RLock()
	defer s.RUnlock()

	return s.resourceVersion
}

/*func (s *ClusterState) GetPods() []v1.Pod {
	s.RLock()
	defer s.RUnlock()

	pods := make([]v1.Pod, len(s.pods))
	i := 0
	for _, v := range s.pods {
		pods[i] = v
		i++
	}

	return pods
}*/

func (s *ClusterState) GetPods(filter PodFilter) []v1.Pod {
	s.RLock()
	defer s.RUnlock()

	pods := make([]v1.Pod, 0)
	for _, pod := range s.pods {
		if filter(&pod) {
			pods = append(pods, pod)
		}
	}

	return pods
}

func (s *ClusterState) GetNodes() []v1.Node {
	s.RLock()
	defer s.RUnlock()

	nodes := make([]v1.Node, len(s.nodes))
	i := 0
	for _, v := range s.nodes {
		nodes[i] = v
		i++
	}

	return nodes
}

func (s *ClusterState) GetPod(name string) (v1.Pod, bool) {
	s.RLock()
	defer s.RUnlock()

	pod, ok := s.pods[name]
	if !ok {
		return v1.Pod{}, false
	}

	return pod, ok
}

func (s *ClusterState) UpdatePod(podName string, newPodState v1.Pod) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.pods[podName]; !ok {
		// TODO: Find out if such situation can happen in our simulation. If yes - fix this one
		panic(fmt.Sprintf("ClusterState UpdatePod error: trying to update a pod with name %s which does not exist!", podName))
	}

	s.pods[podName] = newPodState
}
