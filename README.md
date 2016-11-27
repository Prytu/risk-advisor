# risk-advisor
PoC of risk advisor module for Kubernetes

## Risk advisor server
Usage:

* --port int           Port to listen on (default 11111)
* --proxy-url string   URL of proxy server (default "http://localhost:9998/")

Endpoints:

* /advice	Accepts api.Pod as JSON, returns JSON containing:
 * id: (int) ID of request
 * status: (string) Status of request
 * result: (api.Binding) Information passed from proxy server