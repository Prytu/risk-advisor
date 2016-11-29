package riskadvisor

import "k8s.io/kubernetes/pkg/api"

type AdviceRequest struct {
	Id  int     `json:"id"`
	Pod api.Pod `json:"pod"`
}

type AdviceStatus struct {
	Id     int         `json:"id"`
	Status string      `json:"status"`
	Result api.Binding `json:"result"`
}
