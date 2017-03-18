.PHONY: all deps install test format

all: deps install test

deps:
	@echo "Getting dependencies..."
	@glide install --strip-vendor

install:
	@echo "Installing riskadvisor..."
	@go install ./cmd/riskadvisor
	@echo "Installing simulator..."
	@go install ./cmd/simulator
	@echo "Copying riskadvisor binary"
	@mv ${GOPATH}/bin/riskadvisor docker/risk-advisor/
	@echo "Copying simulator binary"
	@mv ${GOPATH}/bin/simulator docker/simulator/

test:
	@echo "Testing..."
	@go test ./cmd/... ./pkg/...

format:
	@echo "Formatting..."
	@gofmt -w -s cmd pkg 
