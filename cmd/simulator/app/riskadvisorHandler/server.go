package riskadvisorhandler

import (
	"net/http"
)

type RiskAdvisorHandler struct {
	server *http.ServeMux
}

func New(adviseHandler HTTPHandlerFunc) *RiskAdvisorHandler {
	mux := http.NewServeMux()

	mux.HandleFunc("/advise", adviseHandler)

	return &RiskAdvisorHandler{
		server: mux,
	}
}

func (handler *RiskAdvisorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.server.ServeHTTP(w, r)
}
