package riskadvisorhandler

import "k8s.io/kubernetes/pkg/api"

type riskAdvisorRequest struct {
	Pod api.Pod `json:"pod" binding:"required"`
}
