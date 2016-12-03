package main

import (
	"log"
	"net/http"

	flag "github.com/spf13/pflag"

	"github.com/Prytu/risk-advisor/cmd/riskadvisor/app"
	"github.com/Prytu/risk-advisor/pkg/flags"
)

func main() {
	proxyAddress := flag.String("proxy", defaults.ProxyAddress, "Address on which proxy runs")
	port := flag.String("port", defaults.RiskAdvisorUserPort, "Port on which risk-advisors listens for users requests")
	flag.Parse()

	riskAdvisor := app.New(*proxyAddress)

	log.Printf("Starting risk-advisor with:\n\t- port: %v\n\t- proxy URL: %v", *port, *proxyAddress)

	http.ListenAndServe(":"+*port, riskAdvisor)
}
