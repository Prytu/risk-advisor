package riskadvisor

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/kubernetes/pkg/api"
)

type AdviceService struct {
	proxyUrl string
	lastId   int
}

func New(proxyUrl string) http.Handler {
	as := AdviceService{proxyUrl: proxyUrl, lastId: 1}
	wsContainer := restful.NewContainer()
	as.Register(wsContainer)
	return wsContainer
}

func (a *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {
	//TODO: do this atomically
	thisId := a.lastId
	a.lastId++

	pod := new(api.Pod)
	err := request.ReadEntity(&pod)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	ar := AdviceRequest{thisId, *pod}
	arJSON, err := json.Marshal(ar)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp, err := http.Post(a.proxyUrl, "application/json", bytes.NewReader(arJSON))
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}

	bindingJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	binding := api.Binding{}
	err = json.Unmarshal(bindingJSON, &binding)
	if err != nil {
		response.WriteError(http.StatusExpectationFailed, err)
		return
	}

	status := AdviceStatus{thisId, "OK", binding}
	response.WriteEntity(status)
}

func (as *AdviceService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/advise").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("").To(as.sendAdviceRequest).
		// Documentation
		Doc("Post a request for advice").
		Reads(api.Pod{}).
		Returns(200, "OK", AdviceStatus{}))

	container.Add(ws)
}
