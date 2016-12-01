package app

import "k8s.io/kubernetes/pkg/api"

type AdviceResponse struct {
	Status string      `json:"status"`
	Result api.Binding `json:"result"`
}
