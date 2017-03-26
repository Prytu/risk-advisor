package model

import "k8s.io/client-go/1.5/pkg/api/v1"

const MaxNameLength = 58

type SimulatorRequest struct {
	ToCreate []*v1.Pod `json:"toCreate" binding:"required"`
	ToDelete []*v1.Pod `json:"toDelete"`
}

type SchedulingResult struct {
	PodName string `json:"podName"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

type SchedulingError struct {
	ErrorMessage string `json:"errorMessage"`
}
