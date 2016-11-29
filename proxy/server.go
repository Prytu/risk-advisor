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
)

// TODO: create a Binding object instead
const bindingResponse = "{ " +
	"\"kind\": \"Status\"," +
	"\"apiVersion\": \"v1\"," +
	"\"metadata\": {}," +
	"\"status\": \"Success\"," +
	"\"code\": 201}"

type Proxy struct {
	masterURL       *url.URL
	reverseProxy    *httputil.ReverseProxy
	podProvider     podprovider.UnscheduledPodProvider
	responseChannel chan<- api.Binding
	errorChannel    chan<- error
}

func New(serverURL string, podProvider podprovider.UnscheduledPodProvider,
	responseChannel chan<- api.Binding, errorChannel chan<- error) (*Proxy, error) {
	masterURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		masterURL:       masterURL,
		reverseProxy:    httputil.NewSingleHostReverseProxy(masterURL),
		podProvider:     podProvider,
		responseChannel: responseChannel,
		errorChannel:    errorChannel,
	}, nil
}

// TODO: find a better way of error handling. This should probably let client know that an error ocurred before panicing.
// We probably want the panic tho, to spot bugs as soon as possible.
func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if strings.Contains(r.URL.String(), "bindings") {
			log.Printf("POST binding, url = %v", r.URL.String())
			proxy.handleBindings(w, r)
		} else if strings.Contains(r.URL.String(), "events") {
			proxy.handleEvents(w)
		} else {
			panic(fmt.Sprintf("Unexpected POST at URL: %s\n", r.URL.String()))
		}
		return
	}

	if strings.Contains(r.URL.String(), "api/v1/watch/pods") {
		proxy.handleWatches(w, r)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/api/v1/pods") {
		if r.Method == "GET" {
			proxy.handleGetPods(w, r)
			return
		} else {
			panic(fmt.Sprintf("Unexpected Request type: %v at URL: %s\n", r.Method, r.URL.String()))
		}
	}

	proxy.reverseProxy.ServeHTTP(w, r)
}

func (proxy *Proxy) handleBindings(w http.ResponseWriter, r *http.Request) {
	var binding api.Binding

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errWithMessage := fmt.Errorf("Error reading from request body: %v", err)
		proxy.errorChannel <- errWithMessage
		return
	}
	err = json.Unmarshal(body, &binding)
	if err != nil {
		errWithMessage := fmt.Errorf("Error Unmarshalling request body: %v", err)
		proxy.errorChannel <- errWithMessage
		return
	}

	proxy.responseChannel <- binding

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(bindingResponse))
}

func (proxy *Proxy) handleEvents(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte(""))
}

func (proxy *Proxy) handleWatches(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.Write([]byte(""))
}

func (proxy *Proxy) handleGetPods(w http.ResponseWriter, r *http.Request) {
	unescapedURL, err := url.QueryUnescape(r.URL.String())
	if err != nil {
		panic(fmt.Sprintf("Failed to unescape URL: %v", err))
	}

	// here we always respond with an empty pod list
	if strings.Contains(unescapedURL, "spec.nodeName!=") {
		podList := EmptyPodList()

		podJSON, err := json.Marshal(podList)
		if err != nil {
			panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
		}
		log.Printf("GET %s\nResponding with: %s", unescapedURL, string(podJSON))

		w.Header().Set("Content-Type", "application/json")
		w.Write(podJSON)
		return
	}

	var podList *api.PodList

	pod, err := proxy.podProvider.GetPod()
	if err != nil {
		podList = EmptyPodList()
	} else {
		podList = PodListFromPod(&pod)
	}

	podJSON, err := json.Marshal(podList)
	if err != nil {
		panic(fmt.Sprintf("Error marshalling response: %v\n\n", err))
	}

	log.Printf("GET %s\nResponding with: %s", unescapedURL, string(podJSON))

	w.Header().Set("Content-Type", "application/json")
	w.Write(podJSON)
}
