package main

import (
	"log"
	"net/http"

	flag "github.com/spf13/pflag"

	"github.com/Prytu/risk-advisor/cmd/riskadvisor/app"
	"github.com/Prytu/risk-advisor/pkg/flags"
	"github.com/Prytu/risk-advisor/pkg/kubeClient"
)

func main() {
	simulatorPort := flag.String("simulator", defaults.SimulatorPort, "Address on which simulator pod listens for requests")
	port := flag.String("port", defaults.RiskAdvisorUserPort, "Port on which risk-advisors listens for users requests")
	simulatorStartupTimeout := flag.Int("timeout", defaults.Timeout, "Maximum duration in seconds to wait for simulator pod to start running.")
	flag.Parse()

	kubernetesClient, err := kubeClient.New(*simulatorStartupTimeout)
	if err != nil {
		log.Fatalf("Failed to communicate with cluster when building kubeClient: %e\n", err)
	}

	riskAdvisor := app.New(*simulatorPort, kubernetesClient)

	log.Printf("Starting risk-advisor with:\n\t- port: %v\n\t- simulator port: %v", *port, *simulatorPort)

	http.ListenAndServe(":"+*port, riskAdvisor)
}
