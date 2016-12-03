package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/Prytu/risk-advisor/pkg/model"
	"k8s.io/kubernetes/pkg/api"
)

type AdviceService struct {
	proxyUrl string
}

func New(proxyUrl string) http.Handler {
	as := AdviceService{proxyUrl}
	wsContainer := restful.NewContainer()
	as.Register(wsContainer)
	return wsContainer
}

func (as *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {
	var pod api.Pod
	err := request.ReadEntity(&pod)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	ar := model.AdviceRequest{&pod}
	arJSON, err := json.Marshal(ar)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp, err := http.Post(as.proxyUrl+"/advise", "application/json", bytes.NewReader(arJSON))
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}

	bindingJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var binding api.Binding
	err = json.Unmarshal(bindingJSON, &binding)
	if err != nil {
		response.WriteError(http.StatusExpectationFailed, err)
		return
	}

	proxyResponse := AdviceResponse{"OK", binding}
	response.WriteEntity(proxyResponse)
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
		Returns(200, "OK", AdviceResponse{}))

	container.Add(ws)
}
