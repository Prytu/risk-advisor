package app

import (
	"testing"
	"time"
	"net/http"

	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"

	"github.com/stretchr/testify/assert"
	"github.com/emicklei/go-restful"
)

func getDefaultKubernetesClient() kubeClient.ClusterCommunicator {
	kcHttpClient := http.Client{Timeout: time.Duration(defaults.RequestTimeout) * time.Second}
	kubernetesClient, _ := kubeClient.New(kcHttpClient)
	return kubernetesClient
}

func getDefaultRiskAdvisorHttpClient() http.Client {
	return http.Client{Timeout: time.Duration(defaults.RequestTimeout) * time.Second}
}

func getDefaultAdviceService(kubernetesClient kubeClient.ClusterCommunicator, raHttpClient http.Client) AdviceService {
	return GetAdviceService(defaults.SimulatorPort, kubernetesClient, raHttpClient, defaults.StartupTimeout)
}

func getDefaultRegisteredWebServices() []*restful.WebService {
	kubernetesClient := getDefaultKubernetesClient()
	raHttpClient := getDefaultRiskAdvisorHttpClient()
	as := getDefaultAdviceService(kubernetesClient, raHttpClient)
	wsContainer := restful.NewContainer()
	as.Register(wsContainer)
	return wsContainer.RegisteredWebServices()
}

func TestDefaultAdviceService(t *testing.T) {
	kubernetesClient := getDefaultKubernetesClient()
	raHttpClient := getDefaultRiskAdvisorHttpClient()
	as := getDefaultAdviceService(kubernetesClient, raHttpClient)

	assert.Equal(t, as.simulatorPort, defaults.SimulatorPort)
	assert.Equal(t, as.clusterCommunicator, kubernetesClient)
	assert.Equal(t, as.httpClient, raHttpClient)
	assert.Equal(t, as.simulatorStartupTimeout, defaults.StartupTimeout)
}

func TestDefaultRegisteredWebServicesLength(t *testing.T) {
	webServices := getDefaultRegisteredWebServices()
	assert.Equal(t, len(webServices), 1)
}

func TestDefaultRegisteredWebServiceRootPath(t *testing.T) {
	webServices := getDefaultRegisteredWebServices()
	webService := *(webServices[0])
	assert.Equal(t, webService.RootPath(), "/advise")
}

func TestDefaultRegisteredWebServiceRoute(t *testing.T) {
	webServices := getDefaultRegisteredWebServices()
	webService := *(webServices[0])
	routes := webService.Routes()
	assert.Equal(t, len(routes), 1)

	route := routes[0]
	assert.Equal(t, route.Path, "/advise/")
	assert.Equal(t, route.Method, "POST")

	consumes := route.Consumes
	assert.Equal(t, len(consumes), 1)
	assert.Equal(t, consumes[0], restful.MIME_JSON)

	produces := route.Produces
	assert.Equal(t, len(produces), 1)
	assert.Equal(t, produces[0], restful.MIME_JSON)
}
