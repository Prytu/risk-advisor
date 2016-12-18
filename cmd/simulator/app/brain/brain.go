package brain

import (
	"encoding/json"
	"fmt"
	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
	"gopkg.in/gorilla/mux.v1"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TODO: for now we use pod.Name to identify pods. Maybe use uid or something like that instead?
type Brain struct {
	// Snapshot of the state of the cluster
	state *state.ClusterState

	// Channel that will send scheduling events to Simulator
	eventChannel chan<- *v1.Event
}

func New(state *state.ClusterState, eventChannel chan<- *v1.Event) *Brain {
	return &Brain{
		state:        state,
		eventChannel: eventChannel,
	}
}

func (b *Brain) Binding(w http.ResponseWriter, r *http.Request) {
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

	resp := b.handleBinding(&binding)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (b *Brain) Event(w http.ResponseWriter, r *http.Request) {
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

	resp := b.handleEvent(&event)

	w.WriteHeader(http.StatusConflict)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// TODO: Maybe we should return 404 here?
func (b *Brain) GetPod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	podname, ok := vars["podname"]
	if !ok {
		panic("No podname in vars in GetPod")
	}

	pod, ok := b.state.GetPod(podname)
	if !ok {
		panic(fmt.Sprintf("No podname with name %s in state", podname))
	}

	podsJSON, err := json.Marshal(&pod)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podsJSON)
}

func (b *Brain) GetPods(w http.ResponseWriter, r *http.Request) {
	var pods []v1.Pod
	var filter state.PodFilter
	fieldSelector := r.URL.Query().Get("fieldSelector")

	if strings.Contains(fieldSelector, "spec.nodeName!=") {
		filter = func(pod *v1.Pod) bool {
			return pod.Spec.NodeName != "" &&
				pod.Status.Phase != v1.PodSucceeded &&
				pod.Status.Phase != v1.PodFailed
		}
	} else if strings.Contains(fieldSelector, "spec.nodeName=") {
		filter = func(pod *v1.Pod) bool {
			return pod.Spec.NodeName == "" &&
				pod.Status.Phase != v1.PodSucceeded &&
				pod.Status.Phase != v1.PodFailed
		}
	} else {
		filter = func(pod *v1.Pod) bool {
			return true
		}
		log.Printf("Unexpected GET pods field selector: %s", fieldSelector)
	}

	pods = b.state.GetPods(filter)
	resourceVersion := strconv.FormatInt(int64(b.state.GetResourceVersion()), 10)

	podList := v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodList",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			SelfLink:        "/api/v1/pods",
			ResourceVersion: resourceVersion,
		},
		Items: pods,
	}

	podListJSON, err := json.Marshal(&podList)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(podListJSON)
}

func (b *Brain) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes := b.state.GetNodes()
	resourceVersion := strconv.FormatInt(int64(b.state.GetResourceVersion()), 10)

	nodeList := v1.NodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodeList",
			APIVersion: "v1",
		},
		ListMeta: metav1.ListMeta{
			SelfLink:        "/api/v1/nodes",
			ResourceVersion: resourceVersion,
		},
		Items: nodes,
	}

	nodeListJSON, err := json.Marshal(&nodeList)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(nodeListJSON)
}

func (b *Brain) GetPvcs(w http.ResponseWriter, r *http.Request) {
	pvcs := b.state.Pvcs

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvcs)
}

func (b *Brain) GetPvs(w http.ResponseWriter, r *http.Request) {
	pvs := b.state.Pvs

	w.Header().Set("Content-Type", "application/json")
	w.Write(pvs)
}

func (b *Brain) GetReplicasets(w http.ResponseWriter, r *http.Request) {
	replicasets := b.state.Replicasets

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicasets)
}

func (b *Brain) GetServices(w http.ResponseWriter, r *http.Request) {
	services := b.state.Services

	w.Header().Set("Content-Type", "application/json")
	w.Write(services)
}

func (b *Brain) GetReplicationControllers(w http.ResponseWriter, r *http.Request) {
	replicationControllers := b.state.ReplicationControllers

	w.Header().Set("Content-Type", "application/json")
	w.Write(replicationControllers)
}

func (b *Brain) Watchers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func (b *Brain) AddPodToState(pod v1.Pod) {
	resourceVersion := b.state.GetResourceVersion()

	updateNewPodData(&pod, resourceVersion)
	b.state.AddPod(pod)
}

// TODO: maybe generate binding response instead of sending the same for each? (check if it is necessary)
func (b *Brain) handleBinding(binding *v1.Binding) []byte {
	podName := binding.ObjectMeta.Name
	nodeName := binding.Target.Name

	pod, ok := b.state.GetPod(podName)
	if !ok {
		panic(fmt.Sprintf("Error fetching pod from State in Binding handling: pod with name %s not found!", podName))
	}

	bindPodToNode(&pod, nodeName)
	b.state.UpdatePod(podName, pod)

	// here we just bind the pod to node, the scheduling result will be sent as Event and processed there

	return bindingResponse
}

func (b *Brain) handleEvent(event *v1.Event) []byte {
	if event.InvolvedObject.Kind != "Pod" {
		log.Print("Non-pod event.")
		return []byte("")
	}

	// here we send scheduling event
	b.eventChannel <- event

	return []byte("")
}
