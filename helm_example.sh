echo "Starting service"
SVCNAME=`helm install helm-chart/risk-advisor | grep -o -m 1 ".*-risk-advisor"`
echo "Service name: $SVCNAME"
echo "Sending request"
curl -XPOST $(minikube service $SVCNAME --url)/advise -H "Content-type: application/json" -d @pod-examples/pods.json