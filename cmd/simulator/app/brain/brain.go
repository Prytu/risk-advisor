package brain

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"

	"github.com/Prytu/risk-advisor/cmd/simulator/app/state"
)

// TODO: for now we use pod.Name to identify pods. Maybe use uid or something like that instead?
type Brain struct {
	// Snapshot of the state of the cluster
	state *state.ClusterState

	// Channel that will send scheduling events to Simulator
	eventChannel chan<- *v1.Event

	nodesMutex            *sync.Mutex
	isNodesRequestHandled bool
}

func New(state *state.ClusterState, eventChannel chan<- *v1.Event) *Brain {
	mutex := &sync.Mutex{}
	mutex.Lock()

	return &Brain{
		state:                 state,
		eventChannel:          eventChannel,
		nodesMutex:            mutex,
		isNodesRequestHandled: false,
	}
}

func (b *Brain) GetPod(podName string) (*v1.Pod, error) {
	pod, ok := b.state.GetPod(podName)
	if !ok {
		return nil, fmt.Errorf("no podname with name %s in state", podName)
	}

	return &pod, nil
}

func (b *Brain) GetPods(fieldSelector string) *v1.PodList {
	var pods []v1.Pod
	var filter state.PodFilter

	if strings.Contains(fieldSelector, "spec.nodeName!=") {
		filter = state.AssignedNonTerminatedPodFilter
	} else if strings.Contains(fieldSelector, "spec.nodeName=") {
		filter = state.UnassignedNonTerminatedPodFilter
	} else {
		filter = state.AllPodsFilter
		log.Printf("Unexpected GET pods field selector: %s", fieldSelector)
	}

	pods = b.state.GetPods(filter)
	resourceVersion := strconv.FormatInt(int64(b.state.GetResourceVersion()), 10)

	podList := &v1.PodList{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PodList",
			APIVersion: "v1",
		},
		ListMeta: unversioned.ListMeta{
			SelfLink:        "/api/v1/pods",
			ResourceVersion: resourceVersion,
		},
		Items: pods,
	}

	b.nodesMutex.Lock()
	b.nodesMutex.Unlock()
	// Temporary solution: sleep 1 second to make sure response to nodes request is delivered to client
	// TODO: Fix
	time.Sleep(time.Second)

	return podList
}

func (b *Brain) GetNodes() *v1.NodeList {
	nodes := b.state.GetNodes()
	resourceVersion := strconv.FormatInt(int64(b.state.GetResourceVersion()), 10)

	nodeList := &v1.NodeList{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "NodeList",
			APIVersion: "v1",
		},
		ListMeta: unversioned.ListMeta{
			SelfLink:        "/api/v1/nodes",
			ResourceVersion: resourceVersion,
		},
		Items: nodes,
	}

	if !b.isNodesRequestHandled {
		b.isNodesRequestHandled = true
		b.nodesMutex.Unlock()
	}

	return nodeList
}

// TODO: Make a 'generic' function for those functions
func (b *Brain) GetPvcs() *v1.PersistentVolumeClaimList {
	return b.state.Pvcs
}

func (b *Brain) GetPvs() *v1.PersistentVolumeList {
	return b.state.Pvs
}

func (b *Brain) GetReplicasets() *v1beta1.ReplicaSetList {
	return b.state.Replicasets
}

func (b *Brain) GetServices() *v1.ServiceList {
	return b.state.Services
}

func (b *Brain) GetReplicationControllers() *v1.ReplicationControllerList {
	return b.state.ReplicationControllers
}

func (b *Brain) AddPodToState(pod v1.Pod) {
	resourceVersion := b.state.GetResourceVersion()

	updateNewPodData(&pod, resourceVersion)
	b.state.AddPod(pod)
}

func (b *Brain) Watchers() []byte {
	return []byte("")
}

func (b *Brain) Event(event *v1.Event) []byte {
	if event.InvolvedObject.Kind != "Pod" {
		log.Printf("Non-pod event: %v.", event)
		return []byte("")
	}

	// here we send scheduling event
	b.eventChannel <- event

	return []byte("")
}

// TODO: maybe generate binding response instead of sending the same for each? (check if it is necessary)
func (b *Brain) Binding(binding *v1.Binding) ([]byte, error) {
	podName := binding.ObjectMeta.Name
	nodeName := binding.Target.Name

	pod, ok := b.state.GetPod(podName)
	if !ok {
		return nil, fmt.Errorf("Error fetching pod from State in Binding handling: pod with name %s not found!", podName)
	}

	bindPodToNode(&pod, nodeName)
	b.state.UpdatePod(podName, pod)

	// here we just bind the pod to node, the scheduling result will be sent as Event and processed there

	return bindingResponse, nil
}
