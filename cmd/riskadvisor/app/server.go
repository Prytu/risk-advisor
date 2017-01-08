package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/Prytu/risk-advisor/pkg/model"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	metav1 "k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"time"
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

func (as *AdviceService) sendAdviceRequest(request *restful.Request, response *restful.Response) {

	clientset, err := as.getKubernetesClientset()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Creating simulator pod")
	podIp, err := as.createSimulatorPod(clientset)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Getting pods from user request")
	pods, err := as.getPodsFromRequest(request)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Creating simulator request")
	simulatorRequestJSON, err := as.getSimulatorRequestJSON(pods)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Waiting until simulator is ready")
	as.waitUntilSimulatorReady(clientset, podIp)

	log.Printf("Sending simulator request")
	simulatorResponse, err := as.sendSimulatorRequest(podIp, simulatorRequestJSON)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Response received")
	response.WriteEntity(simulatorResponse)

	log.Printf("Deleting simulator pod")
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
	_, err := clientset.CoreV1().Pods("default").Create(&simulatorPod)
	if err != nil {
		return "", err
	}

	newPod, err := clientset.CoreV1().Pods("default").Get("simulator", metav1.GetOptions{})
	for newPod.Status.PodIP == "" {
		time.Sleep(time.Second)
		newPod, err = clientset.CoreV1().Pods("default").Get("simulator", metav1.GetOptions{})
	}
	return newPod.Status.PodIP, nil
}

func (as *AdviceService) getSimulatorAdviseUrl(podIp string) string {
	return "http://" + podIp + ":" + as.simulatorPort + "/advise"
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
	for err != nil {
		resp, err = http.Post(as.getSimulatorAdviseUrl(podIp), "application/json", bytes.NewReader(simulatorRequestJSON))
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
	_ = clientset.CoreV1().Pods("default").Delete("simulator", &v1.DeleteOptions{})
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
