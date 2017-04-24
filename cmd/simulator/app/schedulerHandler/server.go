package schedulerHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/gorilla/mux.v1"
	"k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/brain"
)

type SchedulerHandler struct {
	server  *mux.Router
	Port    string
	brain   *brain.Brain
	errChan chan<- error
}

func New(b *brain.Brain, port string, errChan chan<- error) *SchedulerHandler {
	r := mux.NewRouter()

	sh := &SchedulerHandler{
		server:  r,
		Port:    port,
		brain:   b,
		errChan: errChan,
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

// TODO: Check if we can just 'return' without answering to scheduler in handlers when error happens
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
		sh.handleError(marshallingError("getNodes", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(nodeListJSON)
}

// TODO: Maybe we should return 404 here?
func (sh *SchedulerHandler) getPod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	podName, ok := vars["podname"]
	if !ok {
		sh.handleError(errors.New("No `podname` in vars in getPod."))
		return
	}

	pod, err := sh.brain.GetPod(podName)
	if err != nil {
		sh.handleError(fmt.Errorf("getPod error: %s", err))
	}

	podsJSON, err := json.Marshal(pod)
	if err != nil {
		sh.handleError(marshallingError("getPod", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podsJSON)
}

func (sh *SchedulerHandler) getPods(w http.ResponseWriter, r *http.Request) {
	fieldSelector := r.URL.Query().Get("fieldSelector")

	podList := sh.brain.GetPods(fieldSelector)

	podListJSON, err := json.Marshal(&podList)
	if err != nil {
		sh.handleError(marshallingError("getPods", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podListJSON)
}

func (sh *SchedulerHandler) event(w http.ResponseWriter, r *http.Request) {
	var event v1.Event

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sh.handleError(fmt.Errorf("error reading from request body in event handler: %v", err))
		return
	}

	err = json.Unmarshal(body, &event)
	if err != nil {
		sh.handleError(fmt.Errorf("error unmarshalling request body in event handler: %v", err))
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
		sh.handleError(fmt.Errorf("error reading from request body in binding handler: %v", err))
		return
	}

	err = json.Unmarshal(body, &binding)
	if err != nil {
		sh.handleError(fmt.Errorf("error unmarshalling request body in binding handler: %v", err))
		return
	}

	resp, err := sh.brain.Binding(&binding)
	if err != nil {
		sh.handleError(fmt.Errorf("error in binding handler: %s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// TODO: Make a 'generic' function for those handlers
func (sh *SchedulerHandler) getPvcs(w http.ResponseWriter, r *http.Request) {
	pvcs := sh.brain.GetPvcs()

	pvcsJSON, err := json.Marshal(&pvcs)
	if err != nil {
		sh.handleError(marshallingError("getPvcs", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvcsJSON)
}

func (sh *SchedulerHandler) getPvs(w http.ResponseWriter, r *http.Request) {
	pvs := sh.brain.GetPvs()

	pvsJSON, err := json.Marshal(&pvs)
	if err != nil {
		sh.handleError(marshallingError("getPvs", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvsJSON)
}

func (sh *SchedulerHandler) getReplicasets(w http.ResponseWriter, r *http.Request) {
	replicasets := sh.brain.GetReplicasets()

	replicasetsJSON, err := json.Marshal(&replicasets)
	if err != nil {
		sh.handleError(marshallingError("getReplicasets", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicasetsJSON)
}

func (sh *SchedulerHandler) getServices(w http.ResponseWriter, r *http.Request) {
	services := sh.brain.GetServices()

	servicesJSON, err := json.Marshal(&services)
	if err != nil {
		sh.handleError(marshallingError("getServices", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(servicesJSON)
}

func (sh *SchedulerHandler) getReplicationControllers(w http.ResponseWriter, r *http.Request) {
	replicationControllers := sh.brain.GetReplicationControllers()

	replicationControllersJSON, err := json.Marshal(&replicationControllers)
	if err != nil {
		sh.handleError(marshallingError("getReplicationControllers", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicationControllersJSON)
}

func (sh *SchedulerHandler) handleError(err error) {
	errMsg := fmt.Errorf("SchedulerHandler error: %s", err)
	log.WithError(err).Error(errMsg)
	sh.errChan <- errMsg
}

func marshallingError(handlerName string, err error) error {
	return fmt.Errorf("error marshalling response in %s: %s", handlerName, err)
}
