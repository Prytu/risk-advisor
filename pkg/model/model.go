package model

import "k8s.io/client-go/1.5/pkg/api/v1"

const MaxNameLength = 58

type SimulatorRequest struct {
	ToCreate []*v1.Pod `json:"toCreate" binding:"required"`
	ToDelete []*v1.Pod `json:"toDelete"`
}

type SchedulingResult struct {
	PodName      string `json:"podName,omitempty"`
	Result       string `json:"result,omitempty"`
	Message      string `json:"message,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
