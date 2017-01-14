package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Prytu/risk-advisor/pkg/model"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/rest"
)

var (
	simulatorPod = v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "simulator",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "simulator",
				Image: "pposkrobko/simulator",
				Ports: []v1.ContainerPort{
					{ContainerPort: 9998},
					{ContainerPort: 9999},
				},
			},
				{
					Name:  "kubescheduler",
					Image: "gcr.io/google_containers/kube-scheduler:v1.4.6",
					Command: []string{"/bin/sh", "-c",
						"/usr/local/bin/kube-scheduler --master=127.0.0.1:9999 --leader-elect=false --kube-api-content-type application/json"},
				},
				{
					Name:            "kubectl",
					Image:           "gcr.io/google_containers/kubectl:v0.18.0-120-gaeb4ac55ad12b1-dirty",
					ImagePullPolicy: "Always",
					Args:            []string{"proxy", "-p", "8080"},
				},
			},
		},
	}
)

type AdviceService struct {
	simulatorPort string
}

func New(simulatorPort string) http.Handler {
	as := AdviceService{simulatorPort}
	wsContainer := restful.NewContainer()
	as.Register(wsContainer)
	return wsContainer
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
		Returns(200, "OK", []model.SchedulingResult{}))

	container.Add(ws)
}

func (as *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {

	clientset, err := as.getKubernetesClientset()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Print("Creating simulator pod")
	podIp, err := as.createSimulatorPod(clientset)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Print("Getting pods from user request")
	pods, err := as.getPodsFromRequest(request)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Print("Creating simulator request")
	simulatorRequestJSON, err := as.getSimulatorRequestJSON(pods)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Print("Waiting until simulator is ready")
	as.waitUntilSimulatorReady(clientset, podIp)

	log.Print("Sending simulator request")
	simulatorResponse, err := as.sendSimulatorRequest(podIp, simulatorRequestJSON)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Print("Response received")
	response.WriteEntity(simulatorResponse)

	log.Print("Deleting simulator pod")
	as.deleteSimulatorPod(clientset)
}

func (as *AdviceService) getPodsFromRequest(request *restful.Request) ([]*v1.Pod, error) {

	var pods []*v1.Pod

	/* narazie pazdzierz */
	body, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &pods)
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func (as *AdviceService) getSimulatorRequestJSON(pods []*v1.Pod) ([]byte, error) {

	sr := model.SimulatorRequest{ToCreate: pods}

	srJSON, err := json.Marshal(sr)
	if err != nil {
		return nil, err
	}

	return srJSON, nil
}

func (as *AdviceService) getKubernetesClientset() (*kubernetes.Clientset, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (as *AdviceService) createSimulatorPod(clientset *kubernetes.Clientset) (string, error) {
	_, err := clientset.Core().Pods("default").Create(&simulatorPod)
	if err != nil {
		return "", err
	}

	newPod, err := clientset.Core().Pods("default").Get("simulator")
	for newPod.Status.PodIP == "" {
		time.Sleep(time.Second)
		newPod, err = clientset.Core().Pods("default").Get("simulator")
	}
	return newPod.Status.PodIP, nil
}

func (as *AdviceService) getSimulatorAdviseUrl(podIp string) string {
	return fmt.Sprintf("http://%s:%s/advise", podIp, as.simulatorPort)
}

func (as *AdviceService) waitUntilSimulatorReady(clientset *kubernetes.Clientset, podIp string) {
	// TODO add timeout
	_, err := http.Get(as.getSimulatorAdviseUrl(podIp))
	for err != nil {
		time.Sleep(time.Second)
		_, err = http.Get(as.getSimulatorAdviseUrl(podIp))
	}
}

func (as *AdviceService) sendSimulatorRequest(podIp string, simulatorRequestJSON []byte) ([]model.SchedulingResult, error) {
	resp, err := http.Post(as.getSimulatorAdviseUrl(podIp), "application/json", bytes.NewReader(simulatorRequestJSON))
	if err != nil {
		return nil, err
	}

	responseJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var simulatorResponse []model.SchedulingResult
	err = json.Unmarshal(responseJSON, &simulatorResponse)
	if err != nil {
		return nil, err
	}

	return simulatorResponse, nil
}

func (as *AdviceService) deleteSimulatorPod(clientset *kubernetes.Clientset) {
	// TODO error handling
	_ = clientset.Core().Pods("default").Delete("simulator", &api.DeleteOptions{})
}
