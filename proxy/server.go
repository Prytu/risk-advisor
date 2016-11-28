package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/Prytu/risk-advisor/podprovider"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/api"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"github.com/Prytu/risk-advisor/proxy/mocks"
)

type Proxy struct {
	MasterURL       *url.URL
	ReverseProxy    *httputil.ReverseProxy
	PodProvider     podprovider.UnscheduledPodProvider
	ResponseChannel chan api.Binding
}

func New(serverURL string, podProvider podprovider.UnscheduledPodProvider, responseChannel chan api.Binding) (*Proxy, error) {
	masterURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		MasterURL:       masterURL,
		ReverseProxy:    httputil.NewSingleHostReverseProxy(masterURL),
		PodProvider:     podProvider,
		ResponseChannel: responseChannel,
	}, nil
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.String(), "api/v1/namespaces/default/pods/nginx-without-nodename") {
		log.Printf("NGINX WITHOUT NODENAME: URL: %v, %v", r.URL, r.Method)

		proxy.ReverseProxy.ServeHTTP(w, r)
		return
	}

	if r.Method == "POST" {
		if strings.Contains(r.URL.String(), "bindings") {
			var binding api.Binding

			// TODO: dont panic on errors, push them to an error channel instead
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(fmt.Sprintf("Error reading from request body: %v", err))
			}
			err = json.Unmarshal(body, &binding)
			if err != nil {
				panic(fmt.Sprintf("Error Unmarshalling request body: %v", err))
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
		pod, err := proxy.PodProvider.GetPod()
		if err != nil {
			time.Sleep(5 * time.Second)
			w.Write([]byte(""))
			return
		}
		log.Print("GOT POD, WATCH")

		podEvent := PodEventFromPod(pod)

		eventPodJSON, err := json.MarshalIndent(podEvent, "", "    ")
		if err != nil {
			panic("error marshalling pod event")
			errorMessage := fmt.Sprintf("Error marshalling response: %v\n", err)
			http.Error(w, errorMessage, http.StatusInternalServerError)
		}
		eventPodJSON = append(eventPodJSON, []byte("\r\n")...)

		log.Print(string(eventPodJSON))

		w.Header().Set("Content-Type", "application/json")
		w.Write(eventPodJSON)

		time.Sleep(5 * time.Second)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/api/v1/pods") {
		if r.Method == "GET" {
			var podList *api.PodList

			pod, err := proxy.PodProvider.GetPod()
			if err != nil {
				podList = EmptyPodList()
			} else {
				podList = PodListFromPod(pod)
				log.Print("GOT POD, GET")
			}

			podJSON, err := json.Marshal(podList)
			if err != nil {
				errorMessage := fmt.Sprintf("Error marshalling response: %\n", err)
				http.Error(w, errorMessage, http.StatusInternalServerError)
			}

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
