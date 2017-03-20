package app

import (
	"testing"
	"time"
	"net/http"

	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAdviceService(t *testing.T) {
	simulatorPort := defaults.SimulatorPort
	simulatorStartupTimeout := defaults.StartupTimeout
	simulatorRequestTimeout := defaults.RequestTimeout

	kcHttpClient := http.Client{Timeout: time.Duration(simulatorRequestTimeout) * time.Second}
	kubernetesClient, _ := kubeClient.New(kcHttpClient)
	raHttpClient := http.Client{Timeout: time.Duration(simulatorRequestTimeout) * time.Second}

	as := GetAdviceService(simulatorPort, kubernetesClient, raHttpClient, simulatorStartupTimeout)

	assert.Equal(t, as.simulatorPort, simulatorPort)
	assert.Equal(t, as.clusterCommunicator, kubernetesClient)
	assert.Equal(t, as.httpClient, raHttpClient)
	assert.Equal(t, as.simulatorStartupTimeout, simulatorStartupTimeout)
}
