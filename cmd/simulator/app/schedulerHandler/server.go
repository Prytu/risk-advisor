package schedulerHandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/gorilla/mux.v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
	"k8s.io/client-go/1.5/pkg/api/v1"
)

type SchedulerHandler struct {
	server *mux.Router
	Port   string
	brain  *brain.Brain
}

func New(b *brain.Brain, port string) *SchedulerHandler {
	r := mux.NewRouter()

	sh := &SchedulerHandler{
		server: r,
		Port:   port,
		brain:  b,
	}

	apiv1 := r.PathPrefix("/api/v1/").Subrouter()

	apiv1.HandleFunc("/watch/{resource}", sh.watchers).Methods("GET")

	apiv1.HandleFunc("/nodes", sh.getNodes).Methods("GET")
	apiv1.HandleFunc("/pods", sh.getPods).Methods("GET")
	apiv1.HandleFunc("/persistentvolumeclaims", sh.getPvcs).Methods("GET")
	apiv1.HandleFunc("/persistentvolumes", sh.getPvs).Methods("GET")
	apiv1.HandleFunc("/services", sh.getServices).Methods("GET")
	apiv1.HandleFunc("/replicationcontrollers", sh.getServices).Methods("GET")

	apiv1.HandleFunc("/namespaces/{namespace}/events", sh.event).Methods("POST")

	// TODO: Why do we use the same handler for both of them?
	apiv1.HandleFunc("/namespaces/{namespace}/bindings", sh.binding).Methods("POST")
	apiv1.HandleFunc("/namespaces/{namespace}/pods/{podname}", sh.binding).Methods("POST")

	// PUT for pod update

	extensions := r.PathPrefix("/apis/extensions/v1beta1/").Subrouter()
	extensions.HandleFunc("/replicasets", sh.getReplicasets).Methods("GET")

	return sh
}

func (sh *SchedulerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sh.server.ServeHTTP(w, r)
}

func (sh *SchedulerHandler) watchers(w http.ResponseWriter, r *http.Request) {
	response := sh.brain.Watchers()

	w.Write(response)
}

func (sh *SchedulerHandler) getNodes(w http.ResponseWriter, r *http.Request) {
	nodeList := sh.brain.GetNodes()

	nodeListJSON, err := json.Marshal(&nodeList)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(nodeListJSON)
}

// TODO: Maybe we should return 404 here?
func (sh *SchedulerHandler) getPod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	podName, ok := vars["podname"]
	if !ok {
		panic("No podname in vars in GetPod.")
	}

	pod, err := sh.brain.GetPod(podName)
	if err != nil {
		panic(err)
	}

	podsJSON, err := json.Marshal(pod)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podsJSON)
}

func (sh *SchedulerHandler) getPods(w http.ResponseWriter, r *http.Request) {
	fieldSelector := r.URL.Query().Get("fieldSelector")

	podList := sh.brain.GetPods(fieldSelector)

	podListJSON, err := json.Marshal(&podList)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podListJSON)
}

func (sh *SchedulerHandler) event(w http.ResponseWriter, r *http.Request) {
	var event v1.Event

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading from request body: %v", err))
	}

	err = json.Unmarshal(body, &event)
	if err != nil {
		panic(fmt.Sprintf("Error Unmarshalling request body: %v", err))
		return
	}

	resp := sh.brain.Event(&event)

	w.WriteHeader(http.StatusConflict) // TODO: Check if returning status conflict is 100% OK here
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (sh *SchedulerHandler) binding(w http.ResponseWriter, r *http.Request) {
	var binding v1.Binding

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(fmt.Sprintf("Error reading from request body: %v", err))
	}

	err = json.Unmarshal(body, &binding)
	if err != nil {
		panic(fmt.Sprintf("Error Unmarshalling request body: %v", err))
		return
	}

	resp := sh.brain.Binding(&binding)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// TODO: Make a 'generic' function for those handlers
func (sh *SchedulerHandler) getPvcs(w http.ResponseWriter, r *http.Request) {
	pvcs := sh.brain.GetPvcs()

	pvcsJSON, err := json.Marshal(&pvcs)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvcsJSON)
}

func (sh *SchedulerHandler) getPvs(w http.ResponseWriter, r *http.Request) {
	pvs := sh.brain.GetPvs()

	pvsJSON, err := json.Marshal(&pvs)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvsJSON)
}

func (sh *SchedulerHandler) getReplicasets(w http.ResponseWriter, r *http.Request) {
	replicasets := sh.brain.GetReplicasets()

	replicasetsJSON, err := json.Marshal(&replicasets)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicasetsJSON)
}

func (sh *SchedulerHandler) getServices(w http.ResponseWriter, r *http.Request) {
	services := sh.brain.GetServices()

	servicesJSON, err := json.Marshal(&services)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(servicesJSON)
}

func (sh *SchedulerHandler) getReplicationControllers(w http.ResponseWriter, r *http.Request) {
	replicationControllers := sh.brain.GetReplicationControllers()

	replicationControllersJSON, err := json.Marshal(&replicationControllers)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v.", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicationControllersJSON)
}
