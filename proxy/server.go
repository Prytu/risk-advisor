package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	MasterURL    *url.URL
	ReverseProxy *httputil.ReverseProxy
	podProvider  UnscheduledPodProvider
}

func New(serverURL string, podProvider UnscheduledPodProvider) (*Proxy, error) {
	masterURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		MasterURL:    masterURL,
		ReverseProxy: httputil.NewSingleHostReverseProxy(masterURL),
		podProvider:  podProvider,
	}, nil
}

const bindingResponse = "{ " +
	"\"kind\": \"Status\"," +
	"\"apiVersion\": \"v1\"," +
	"\"metadata\": {}," +
	"\"status\": \"Success\"," +
	"\"code\": 201}"

// this response should be based on scheduler request and it should just add
// selfLink, uid, resourceVersion and creatioTimestamp fields
const eventResponse = "{" +
	"\"kind\": \"Event\"," +
	"\"apiVersion\": \"v1\"," +
	"\"metadata\": {" +
	"\"name\": \"nginx-without-nodename.148a9fcaa3a27080\"," +
	"\"namespace\": \"default\"," +
	"\"selfLink\": \"/api/v1/namespaces/default/events/nginx-without-nodename.148a9fcaa3a27080\"," +
	"\"uid\": \"851920f3-b3e6-11e6-9514-000c2999b232\"," +
	"\"resourceVersion\": \"114\"," +
	"\"creationTimestamp\": \"2016-11-26T14:42:13Z\"}," +
	"\"involvedObject\": {\"kind\": \"Pod\"," +
	"\"namespace\": \"default\"," +
	"\"name\": \"nginx-without-nodename\"," +
	"\"uid\": \"b0220242-b346-11e6-a633-000c2999b232\"," +
	"\"apiVersion\": \"v1\"," +
	"\"resourceVersion\": \"614\"}," +
	"\"reason\": \"Scheduled\"," +
	"\"message\": \"Successfully assigned nginx-without-nodename to ubuntu\"," +
	"\"source\": {" +
	"\"component\": \"default-scheduler\"}," +
	"\"firstTimestamp\": \"2016-11-26T14:38:40Z\"," +
	"\"lastTimestamp\": \"2016-11-26T14:38:40Z\"," +
	"\"count\": 1," +
	"\"type\": \"Normal\"}"

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL: %v, %v", r.URL, r.Method)

	if r.Method == "PUT" {
		log.Printf("PUT content-type: %s, accept: %v\n", r.Header["Content-Type"], r.Header["Accept"])
		buf, _ := ioutil.ReadAll(r.Body)
		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

		body, _ := ioutil.ReadAll(rdr1)
		log.Printf("BODY: %s\n", string(body))

		r.Body = rdr2
	}
	if r.Method == "POST" {
		log.Printf("PUT content-type: %s, accept: %v\n", r.Header["Content-Type"], r.Header["Accept"])

		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("BODY: %s\n", string(body))

		if strings.Contains(r.URL.String(), "bindings") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(bindingResponse))
		} else if strings.Contains(r.URL.String(), "events") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(eventResponse))
		}

		return
	}

	if strings.HasPrefix(r.URL.String(), "/api/v1/pods") {
		log.Printf("URL: %v, %v, %v", r.URL, r.Method, r.Header["Accept"])

		if r.Method == "GET" {
			podList := proxy.podProvider.ProvidePods()

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
		}
	}

	proxy.ReverseProxy.ServeHTTP(w, r)
}
