package proxy

import "k8s.io/kubernetes/pkg/api"

type PodEvent struct {
	Type   string    `json:"type"`
	Object *EventPod `json:"object"`
}

type EventPod struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	*api.Pod   `json:",inline"`
}

func PodEventFromPod(pod *api.Pod) *PodEvent {
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
