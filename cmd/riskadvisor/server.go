package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"log"

	"github.com/emicklei/go-restful"


	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"fmt"
	"time"
)

type AdviceService struct {
	proxyUrl string
}

type SimulatorRequest struct {
	ToCreate []*v1.Pod `json:"toCreate" binding:"required"`
	ToDelete []*v1.Pod `json:"toDelete"`
}

type SchedulingResult struct {
	PodName string `json:"podName"`
	Result  string `json:"result"`
	Message string `json:"message"`
}

func New(proxyUrl string) http.Handler {
	as := AdviceService{proxyUrl}
	wsContainer := restful.NewContainer()
	as.Register(wsContainer)
	return wsContainer
}

func (as *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {
	log.Println("zaczynamy obsluge")

	var pods []*v1.Pod
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
	sr := SimulatorRequest{ToCreate: pods}
log.Println("simulator request")
	srJSON, err := json.Marshal(sr)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
log.Println("simulator request json")
	//var podJson = "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"simulator\"},\"spec\":{\"containers\":[{\"name\":\"simulator\",\"image\":\"pposkrobko/simulator\"}]}}"

	var simulatorPod v1.Pod
	simulatorPod = v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "simulator",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{v1.Container{
					Name: "simulator",
					Image: "pposkrobko/simulator",
					Ports: []v1.ContainerPort{v1.ContainerPort{ContainerPort: 9998},v1.ContainerPort{ContainerPort: 9999},},
				},
				v1.Container{
						Name: "kubescheduler",
						Image: "gcr.io/google_containers/kube-scheduler:v1.4.6",
						Command: []string{"/bin/sh", "-c",  "/usr/local/bin/kube-scheduler --master=127.0.0.1:9999 --leader-elect=false --kube-api-content-type application/json"},
				},
				v1.Container{
						Name: "kubectl",
						Image: "gcr.io/google_containers/kubectl:v0.18.0-120-gaeb4ac55ad12b1-dirty",
						ImagePullPolicy: "Always",
						Args: []string{"proxy", "-p", "8080"},
				},
			},
		},
	}

	// uses the current context in kubeconfig
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
	resp, err := http.Get("http://" + podIp + ":9998/advise")
	for err != nil {
fmt.Println(err)
newPod, err = clientset.CoreV1().Pods("default").Get("simulator", metav1.GetOptions{})
podIp = newPod.Status.PodIP
		time.Sleep(time.Second)
		resp, err = http.Get("http://" + podIp + ":9998/advise")
	}
log.Printf("asking simulator")
	resp, err = http.Post("http://" + podIp + ":9998/advise", "application/json", bytes.NewReader(srJSON))
	for err != nil {
		fmt.Println(err)
		resp, err = http.Post("http://" + podIp + ":9998/advise", "application/json", bytes.NewReader(srJSON))
	}
log.Printf("asked simulator successfully")
	responseJSON, err := ioutil.ReadAll(resp.Body)
log.Printf(string(responseJSON))
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}
log.Printf("got response from simulator")
	var simulatorResponse []SchedulingResult
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
		Reads([]v1.Pod{}).
		Returns(200, "OK", AdviceResponse{}))

	container.Add(ws)
}