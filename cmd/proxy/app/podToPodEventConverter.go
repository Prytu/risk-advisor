package app

import "k8s.io/kubernetes/pkg/api/v1"

// TODO: Use models from kubernetes instead
type PodEvent struct {
	Type   string    `json:"type"`
	Object *EventPod `json:"object"`
}

type EventPod struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	*v1.Pod    `json:",inline"`
}

// This will be used in future in watchers
func PodEventFromPod(pod *v1.Pod) *PodEvent {
	EventPod := &EventPod{
		Kind:       "Pod",
		ApiVersion: "v1",
		Pod:        pod,
	}

	return &PodEvent{
		Type:   "ADDED",
		Object: EventPod,
	}
}
