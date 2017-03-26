package riskadvisorhandler

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/gorilla/mux.v1"
)

type RiskAdvisorHandler struct {
	server *mux.Router
}

func New(adviseHandler HTTPHandlerFunc) *RiskAdvisorHandler {
	r := mux.NewRouter()

	r.HandleFunc("/advise", adviseHandler).Methods("POST")
	r.HandleFunc("/advise", aliveHandler).Methods("GET")

	return &RiskAdvisorHandler{
		server: r,
	}
}

func (handler *RiskAdvisorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.server.ServeHTTP(w, r)
}

func aliveHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Responding to risk-advisor alive check.")
	w.Write([]byte(""))
}
