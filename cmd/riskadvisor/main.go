package main

import (
	"log"
	"net/http"

	flag "github.com/spf13/pflag"

	"github.com/Prytu/risk-advisor/cmd/riskadvisor/app"
	"github.com/Prytu/risk-advisor/pkg/flags"
)

func main() {
	simulatorPort := flag.String("simulator", defaults.SimulatorPort, "Address on which simulator pod listens for requests")
	port := flag.String("port", defaults.RiskAdvisorUserPort, "Port on which risk-advisors listens for users requests")
	flag.Parse()

	riskAdvisor := app.New(*simulatorPort)

	log.Printf("Starting risk-advisor with:\n\t- port: %v\n\t- simulator port: %v", *port, *simulatorPort)

	http.ListenAndServe(":"+*port, riskAdvisor)
}
