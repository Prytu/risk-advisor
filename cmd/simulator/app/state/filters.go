package state

import "k8s.io/client-go/pkg/api/v1"

type PodFilter func(pod *v1.Pod) bool

var AllPodsFilter = func(pod *v1.Pod) bool {
	return true
}

var AssignedNonTerminatedPodFilter = func(pod *v1.Pod) bool {
	return pod.Spec.NodeName != "" &&
		pod.Status.Phase != v1.PodSucceeded &&
		pod.Status.Phase != v1.PodFailed
}

var UnassignedNonTerminatedPodFilter = func(pod *v1.Pod) bool {
	return pod.Spec.NodeName == "" &&
		pod.Status.Phase != v1.PodSucceeded &&
		pod.Status.Phase != v1.PodFailed
}
