FROM golang

ADD $GOPATH/src/github.com/Prytu/risk-advisor ./go/src/github.com/Prytu/risk-advisor/

RUN go install github.com/Prytu/risk-advisor/cmd/riskadvisor

ENTRYPOINT /go/bin/riskadvisor

EXPOSE 9997 
