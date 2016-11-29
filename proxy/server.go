package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/api"

	"github.com/Prytu/risk-advisor/podprovider"
	"github.com/Prytu/risk-advisor/proxy/mocks"
)

type Proxy struct {
	MasterURL       *url.URL
	ReverseProxy    *httputil.ReverseProxy
	PodProvider     podprovider.UnscheduledPodProvider
	ResponseChannel chan<- api.Binding
	ErrorChannel    chan<- error
}

func New(serverURL string, podProvider podprovider.UnscheduledPodProvider,
	responseChannel chan<- api.Binding, errorChannel chan<- error) (*Proxy, error) {
	masterURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		MasterURL:       masterURL,
		ReverseProxy:    httputil.NewSingleHostReverseProxy(masterURL),
		PodProvider:     podProvider,
		ResponseChannel: responseChannel,
		ErrorChannel:    errorChannel,
	}, nil
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if strings.Contains(r.URL.String(), "bindings") {
			var binding api.Binding

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errWithMessage := fmt.Errorf("Error reading from request body: %v", err)
				proxy.ErrorChannel <- errWithMessage
				return
			}
			err = json.Unmarshal(body, &binding)
			if err != nil {
				errWithMessage := fmt.Errorf("Error Unmarshalling request body: %v", err)
				proxy.ErrorChannel <- errWithMessage
				return
			}

			proxy.ResponseChannel <- binding

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(mocks.BindingResponse))
		} else if strings.Contains(r.URL.String(), "events") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(mocks.EventResponse))
		}

		return
	}

	if strings.Contains(r.URL.String(), "api/v1/watch/pods") {
		//pod, err := proxy.PodProvider.GetPod()
		//if err != nil {
		time.Sleep(5 * time.Second)
		w.Write([]byte(""))
		return
		//}
		/*
			podEvent := PodEventFromPod(&pod)

			flusher, ok := w.(http.Flusher)
			if !ok {
				panic("writer is not a flusher")
			}

			eventPodJSON, err := json.MarshalIndent(podEvent, "", "    ")
			if err != nil {
				panic("Error marshalling pod event.")
			}
			eventPodJSON = append(eventPodJSON, []byte("\r\n")...)

			log.Print(string(eventPodJSON))

			w.Header().Set("Content-Type", "application/json")
			w.Write(eventPodJSON)
			flusher.Flush()*/

		return
	}

	// TODO: find a better way of error handling. This should probably let client know that an error ocurred before
	// panicing.
	if strings.HasPrefix(r.URL.String(), "/api/v1/pods") {
		if r.Method == "GET" {
			values := r.URL.Query()

			for k, v := range values {
				//log.Printf("k = %v, v = %v", k, v)
				if k == "fieldSelector" {
					for _, selector := range v {
						if strings.Contains(selector, "spec.nodeName!=") {
							podList := EmptyPodList()

							podJSON, err := json.Marshal(podList)
							if err != nil {
								panic(fmt.Sprintf("Error marshalling response: %v\n", err))
							}

							log.Printf("GET responding with: %s", string(podJSON))

							w.Header().Set("Content-Type", "application/json")
							w.Write(podJSON)
							return
						}
					}
				}
			}

			var podList *api.PodList

			pod, err := proxy.PodProvider.GetPod()
			if err != nil {
				podList = EmptyPodList()
			} else {
				podList = PodListFromPod(&pod)
			}

			podJSON, err := json.Marshal(podList)
			if err != nil {
				panic(fmt.Sprintf("Error marshalling response: %v\n", err))
			}

			log.Printf("GET responding with: %s", string(podJSON))

			w.Header().Set("Content-Type", "application/json")
			w.Write(podJSON)
			return
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
	}

	proxy.ReverseProxy.ServeHTTP(w, r)
}
