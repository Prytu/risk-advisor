package model

import "k8s.io/kubernetes/pkg/api"

type AdviceRequest struct {
	Pod *api.Pod `json:"pod"`
}
