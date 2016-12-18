package schedulerHandler

import (
	"net/http"

	"gopkg.in/gorilla/mux.v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
)

type SchedulerHandler struct {
	server *mux.Router
	Port   string
}

func New(b *brain.Brain, port string) *SchedulerHandler {
	r := mux.NewRouter()

	apiv1 := r.PathPrefix("/api/v1/").Subrouter()

	apiv1.HandleFunc("/watch/{resource}", b.Watchers).Methods("GET")

	apiv1.HandleFunc("/nodes", b.GetNodes).Methods("GET")
	apiv1.HandleFunc("/pods", b.GetPods).Methods("GET")
	apiv1.HandleFunc("/persistentvolumeclaims", b.GetPvcs).Methods("GET")
	apiv1.HandleFunc("/persistentvolumes", b.GetPvcs).Methods("GET")
	apiv1.HandleFunc("/services", b.GetServices).Methods("GET")
	apiv1.HandleFunc("/replicationcontrollers", b.GetServices).Methods("GET")

	apiv1.HandleFunc("/namespaces/{namespace}/events", b.Event).Methods("POST")
	apiv1.HandleFunc("/namespaces/{namespace}/bindings", b.Binding).Methods("POST")
	apiv1.HandleFunc("/namespaces/{namespace}/pods/{podname}", b.Binding).Methods("POST")

	// PUT for pod update

	extensions := r.PathPrefix("/apis/extensions/v1beta1/").Subrouter()
	extensions.HandleFunc("/replicasets", b.GetReplicasets).Methods("GET")

	return &SchedulerHandler{
		server: r,
		Port:   port,
	}
}

func (s *SchedulerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}
