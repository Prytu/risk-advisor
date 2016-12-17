package watcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/kubernetes/pkg/runtime"
)

// EventType defines the possible types of events.
type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Error    EventType = "ERROR"
)

// Event represents a single event to a watched resource.
//TODO: use versioned.Event instead - it uses runtime.RawExtension instead of runtime.Object, not sure how it works.
type Event struct {
	Type EventType `json:"type"`

	// Object is:
	//  * If Type is Added or Modified: the new state of the object.
	//  * If Type is Deleted: the state of the object immediately before deletion.
	//  * If Type is Error: *api.Status is recommended; other types may make sense
	//    depending on context.
	Object runtime.Object `json:"object"`
}

type Watcher struct {
	result  chan Event
	Stopped bool
	sync.Mutex
}

func NewWatcher() *Watcher {
	return &Watcher{
		result: make(chan Event),
	}
}

func (f *Watcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		close(f.result)
		f.Stopped = true
	}
}

func (f *Watcher) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.Stopped
}

// Reset prepares the watcher to be reused.
func (f *Watcher) Reset() {
	f.Lock()
	defer f.Unlock()
	f.Stopped = false
	f.result = make(chan Event)
}

func (f *Watcher) ResultChan() <-chan Event {
	return f.result
}

// Add sends an add event.
func (f *Watcher) Add(obj runtime.Object) {
	f.Action(Added, obj)
}

// Modify sends a modify event.
func (f *Watcher) Modify(obj runtime.Object) {
	f.Action(Modified, obj)
}

// Delete sends a delete event.
func (f *Watcher) Delete(lastValue runtime.Object) {
	f.Action(Deleted, lastValue)
}

// Error sends an Error event.
func (f *Watcher) Error(errValue runtime.Object) {
	f.Action(Error, errValue)
}

// Action sends an event of the requested type
func (f *Watcher) Action(action EventType, obj runtime.Object) {
	select {
	case f.result <- Event{Type: action, Object: obj}:

	default:
		fmt.Println("watcher: no listeners, event ignored")
	}
}

func (watcher *Watcher) ServeHTTP(w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	cn, ok := w.(http.CloseNotifier)
	if !ok {
		panic(fmt.Sprint("Failed to get http.CloseNotifier"))
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		panic(fmt.Sprint("Failed to get http.Flusher"))
	}

	timer := time.NewTimer(timeout)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := watcher.ResultChan()
	for {
		select {
		case <-cn.CloseNotify():
			fmt.Println("DEBUG: watch closed by CloseNotify")
			return
		case <-timer.C:
			fmt.Println("DEBUG: watch closed by timeout")
			return
		case event, ok := <-ch:
			if !ok {
				// End of results.
				return
			}

			eventJSON, err := json.Marshal(event)
			if err != nil {
				panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
			}

			w.Write(eventJSON)
			fmt.Printf("Sending event >>%s<<\n", eventJSON)

			if len(ch) == 0 {
				flusher.Flush()
			}
		}
	}
}
