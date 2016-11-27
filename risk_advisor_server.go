package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/api"
)

type AdviceRequest struct {
	Id  int     `json:"id"`
	Pod api.Pod `json:"pod"`
}

type AdviceStatus struct {
	Id     int         `json:"id"`
	Status string      `json:"status"`
	Result api.Binding `json:"result"`
}

type AdviceService struct {
	proxyUrl string
	lastId   int
}

func (a *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {
	//TODO: do this atomically
	thisId := a.lastId
	a.lastId++

	pod := new(api.Pod)
	err := request.ReadEntity(&pod)
	if err == nil {
		ar := AdviceRequest{thisId, *pod}
		ar_json, err := json.Marshal(ar)
		if err != nil {
			response.WriteError(http.StatusInternalServerError, err)
			return
		}

		resp, err := http.Post(a.proxyUrl, "application/json", bytes.NewReader(ar_json))
		if err != nil {
			response.WriteError(http.StatusNotFound, err)
			return
		}

		binding_json, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			response.WriteError(http.StatusInternalServerError, err)
			return
		}

		binding := api.Binding{}
		err = json.Unmarshal(binding_json, &binding)
		if err != nil {
			response.WriteError(http.StatusExpectationFailed, err)
			return
		}

		status := AdviceStatus{thisId, "OK", binding}
		response.WriteEntity(status)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func main() {
	var port *int = flag.Int("port", 11111, "Port to listen on")
	var proxyUrl *string = flag.String("proxy-url", "http://localhost:9998/", "URL of proxy server")
	flag.Parse()

	as := AdviceService{proxyUrl: *proxyUrl, lastId: 1}
	aws := new(restful.WebService)
	aws.Path("/advice").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	aws.Route(aws.POST("").To(as.sendAdviceRequest).
		// Documentation
		Doc("Post a request for advice").
		Reads(api.Pod{}).
		Returns(200, "OK", AdviceStatus{}))

	restful.Add(aws)

	log.Printf("start listening on localhost:%d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
