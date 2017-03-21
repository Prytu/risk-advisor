package main

import (
	"net/http"
	"time"
	"fmt"

	flag "github.com/spf13/pflag"

	"github.com/Prytu/risk-advisor/cmd/riskadvisor/app"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
	log "github.com/Sirupsen/logrus"
)

func main() {
	simulatorPort := flag.String("simulator", defaults.SimulatorPort, "Address on which simulator pod listens for requests")
	port := flag.String("port", defaults.RiskAdvisorUserPort, "Port on which risk-advisors listens for users requests")
	simulatorStartupTimeout := flag.Int("startupTimeout", defaults.StartupTimeout, "Maximum duration in seconds to wait for simulator pod to start running.")
	simulatorRequestTimeout := flag.Int("requestTimeout", defaults.RequestTimeout, "Maximum duration in seconds to wait for simulator to respond to schedluing request.")

	flag.Parse()

	kcHttpClient := http.Client{Timeout: time.Duration(*simulatorRequestTimeout) * time.Second}
	kubernetesClient, err := kubeClient.New(kcHttpClient)
	if err != nil {
		log.Fatalf("Failed to communicate with cluster when building kubeClient: %e\n", err)
	}

	raHttpCient := http.Client{Timeout: time.Duration(*simulatorRequestTimeout) * time.Second}
	riskAdvisor := app.New(*simulatorPort, kubernetesClient, raHttpCient, *simulatorStartupTimeout)

	log.Printf("Starting risk-advisor with:\n\t- port: %v\n\t- simulator port: %v", *port, *simulatorPort)

	http.ListenAndServe(fmt.Sprintf(":%s", *port), riskAdvisor)
}
