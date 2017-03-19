#!/bin/bash

echo -e "\n### Deleting risk-advisor and risk-advisor-service from Minikube"
kubectl delete pods,services -l name=risk-advisor --grace-period=0 --force

echo -e "\n### Starting risk-advisor pod"
kubectl create -f "docker/pod.yaml"

echo -e "\n### Starting risk-advisor service"
kubectl create -f "docker/service.yaml"

echo -e "\n### Running example request"
curl -XPOST $(minikube service risk-advisor-service --url)/advise -H "Content-type: application/json" -d @pod-examples/pods.json
