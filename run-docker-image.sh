#!/bin/bash

echo "running risk advisor docker image locally"
echo "sudo docker run --publish 6060:11111 --name risk-advisor --rm risk-advisor"
sudo docker run --publish 6060:11111 --name risk-advisor --rm risk-advisor
