package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"log"

	"github.com/emicklei/go-restful"

	"github.com/Prytu/risk-advisor/pkg/model"
	kubeapi "k8s.io/kubernetes/pkg/api/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	clientapi "k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"fmt"
	"time"
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
	log.Println("zaczynamy obsluge")

	var pods []*kubeapi.Pod
	//err := request.ReadEntity(pod)

	/* narazie pazdzierz */
	body, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Println("wczytane do body")
	err = json.Unmarshal(body, &pods)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
	/* koniec pazdzierza */
	log.Println("zjosonowane")
	sr := model.SimulatorRequest{ToCreate: pods}
log.Println("simulator request")
	srJSON, err := json.Marshal(sr)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
log.Println("simulator request json")
	var simulatorPod clientapi.Pod
	simulatorPod = clientapi.Pod{
		ObjectMeta: clientapi.ObjectMeta{
			Name: "simulator",
		},
		Spec: clientapi.PodSpec{
			Containers: []clientapi.Container{clientapi.Container{
				Name: "simulator",
				Image: "pposkrobko/simulator",
				Ports: []clientapi.ContainerPort{clientapi.ContainerPort{ContainerPort: 9997},},
			},},
		},
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
log.Println("mamy in cluster config")

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
log.Printf("mamy clientset")
	_, err = clientset.CoreV1().Pods("default").Create(&simulatorPod)
	if err != nil {
		fmt.Printf(err.Error())
	}
log.Printf("created pod")
	newPod, err := clientset.CoreV1().Pods("default").Get("simulator", metav1.GetOptions{})
	for err != nil {
		time.Sleep(time.Second)
		newPod, err = clientset.CoreV1().Pods("default").Get("simulator", metav1.GetOptions{})
	}
log.Printf("got ip")
	var podIp = newPod.Status.PodIP
	resp, err := http.Post(
		"http://" + podIp + ":9998/advise",
		"application/json",
		bytes.NewReader(srJSON),
	)
	for err != nil {
		time.Sleep(time.Second)
		resp, err = http.Get("http://" + podIp + ":9998/advise")
	}
log.Printf("asking simulator")
	resp, err = http.Post(as.proxyUrl+"/advise", "application/json", bytes.NewReader(srJSON))
	if err != nil {
		response.WriteError(http.StatusNotFound, err)
		return
	}
log.Printf("asked simulator")
	responseJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
log.Printf("got response from simulator")
	var simulatorResponse []model.SchedulingResult
	err = json.Unmarshal(responseJSON, &simulatorResponse)
	if err != nil {
		response.WriteError(http.StatusExpectationFailed, err)
		return
	}
log.Printf("unmarhsalled response")
	response.WriteEntity(simulatorResponse)
}

func (as *AdviceService) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/advise").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("").To(as.sendAdviceRequest).
		// Documentation
		Doc("Post a request for advice").
		Reads([]kubeapi.Pod{}).
		Returns(200, "OK", AdviceResponse{}))

	container.Add(ws)
}
