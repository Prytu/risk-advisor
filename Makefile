.PHONY: all deps install test docker docker-push docker-full format clean check-DOCKER_HUB_USER

all: clean deps install test

deps:
	@echo "Getting dependencies..."
	@glide install --strip-vendor

install:
	@echo "Installing risk-advisor..."
	@go install ./cmd/riskadvisor
	@echo "Installing simulator..."
	@go install ./cmd/simulator
	@echo "Copying risk-advisor binary"
	@mv ${GOPATH}/bin/riskadvisor docker/risk-advisor/
	@echo "Copying simulator binary"
	@mv ${GOPATH}/bin/simulator docker/simulator/

docker:
	@echo "Building risk-advisor docker image";
	@cd docker/risk-advisor; ./build.sh >/dev/null
	@echo "Building simulator docker image";
	@cd docker/simulator; ./build.sh >/dev/null

docker-push: check-DOCKER_HUB_USER
	@echo "Tagging risk-advisor image"
	docker tag $(shell docker images | grep risk-advisor | awk '{ print $$3 }' | head -1) "$(DOCKER_HUB_USER)/risk-advisor:latest" >/dev/null
	@echo "Pushing risk-advisor image"
	@docker push "$(DOCKER_HUB_USER)/risk-advisor:latest" >/dev/null
	@echo "Tagging simulator image"
	@docker tag $(shell docker images | grep simulator| awk '{ print $$3 }' | head -1) "$(DOCKER_HUB_USER)/simulator:latest" >/dev/null
	@echo "Pushing simulator image"
	@docker push "$(DOCKER_HUB_USER)/simulator:latest" >/dev/null
	@echo "Successfuly pushed images"

docker-full: clean install docker docker-push

test:
	@echo "Testing..."
	@go test ./cmd/... ./pkg/...

format:
	@echo "Formatting..."
	@gofmt -w -s cmd pkg 

clean:
	@echo "Deleting riskadvisor binary"
	-@rm docker/risk-advisor/riskadvisor
	@echo "Deleting simulator binary"
	-@rm docker/simulator/simulator
	@echo "Removing old risk-advisor docker image"
	-@docker rmi -f $(shell docker images | grep risk-advisor | awk '{ print $$3 }') >/dev/null 2>/dev/null
	@echo "Removing old simulator docker image"
	-@docker rmi -f $(shell docker images | grep simulator | awk '{ print $$3 }') >/dev/null 2>/dev/null

check-DOCKER_HUB_USER:
ifndef DOCKER_HUB_USER
	$(error DOCKER_HUB_USER has to be defined in order to push images.)
endif