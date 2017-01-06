package main

import (
	"fmt"
	"net/http"
	"log"
	"net"
	"k8s.io/kubernetes/pkg/genericapiserver"
)

func custom(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if (req.URL.String() == "/hello") {
			w.Write([]byte("world"))
			return
		}
		handler.ServeHTTP(w, req)
	})
}

func CustomBuildHandlerChain(apiHandler http.Handler, c *genericapiserver.Config) (secure, insecure http.Handler) {
	secureHandler, insecureHandler := genericapiserver.DefaultBuildHandlerChain(apiHandler, c)
	return secureHandler, custom(insecureHandler)
}

func main() {
	var config = genericapiserver.NewConfig()

	var publicAddress = net.IPv4(127, 0, 0, 1)
	config.PublicAddress = publicAddress

	var insecureServingInfo = &genericapiserver.ServingInfo {
		BindAddress: "127.0.0.1:6442",
		BindNetwork: "tcp",
	}
	config.InsecureServingInfo = insecureServingInfo

	config.BuildHandlerChainsFunc = CustomBuildHandlerChain

	var completedConfig = config.Complete()

	genericServer, err := completedConfig.New()
	if (err != nil) {
		log.Fatal(err)
	}

	stopChannel := make(chan struct{})
	genericServer.PrepareRun().Run(stopChannel)

	fmt.Println("stop")
}
