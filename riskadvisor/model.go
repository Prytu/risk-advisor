package riskadvisor

import "k8s.io/kubernetes/pkg/api"

type AdviceRequest struct {
	Pod *api.Pod `json:"pod"`
}

type AdviceResponse struct {
	Status string      `json:"status"`
	Result api.Binding `json:"result"`
}
