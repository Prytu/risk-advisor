package model

import "k8s.io/kubernetes/pkg/api/v1"

type AdviceRequest struct {
	Pod *v1.Pod `json:"pod"`
}

type ProxyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type FailedSchedulingResponse struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
