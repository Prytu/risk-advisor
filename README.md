# risk-advisor
PoC of risk advisor module for Kubernetes. This project is licensed under the terms of the Apache 2.0 license.

## Risk advisor server
Usage:

* `--port` int                      Port to listen on (default 11111)
* `--proxy-url` string              URL of proxy server (default "http://localhost:9998/")
* `--simulatorStartupTimeout` int   Maximum ammount of time in seconds to wait for simulator pod to start running (default 30)
* `--simulatorRequestTimeout` int   Maximum ammount of time in seconds to wait for simulator to respond to request (default 10)

Endpoints:

 * `/advice`	Accepts api.Pod as JSON, returns JSON containing:
     * `id`: (int) ID of request
     * `status`: (string) Status of request
     * `result`: (api.Binding) Information passed from proxy server

## Building
* `make clean` deletes executables and removes all `risk-advisor` and `simulator` docker images
* `make install` builds executables
* `make docker` builds docker images
* `make docker-tag` tags both docker images as `$(DOCKER_HUB_USER)/(risk-advisor|simulator):latest`
* `make docker-full` performs all above operations
* `make docker-push` pushes `$(DOCKER_HUB_USER)/(risk-advisor|simulator):latest` images to docker hub

If you are developing on macOS you need to `install` on a linux machine. Then you can run `make docker && make docker-tag`
to create and tag docker images.

<b>NOTE</b> that `make docker-tag`, `make docker-push` and `make docker-full` commands require you to have `DOCKER_HUB_USER`
env variable set to your docker hub username. You can also run those commands providing the username like this:
`<command> DOCKER_HUB_USERNAME=<username>`.
