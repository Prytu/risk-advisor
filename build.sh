#!/bin/bash

go install ./cmd/riskadvisor
go install ./cmd/simulator

mv $GOPATH/bin/riskadvisor docker/risk-advisor/
mv $GOPATH/bin/simulator docker/simulator/
