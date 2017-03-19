#!/bin/bash

echo "Building project"
(cd ..; ./build.sh)

echo "Removing old risk-advisor docker image"
docker rmi -f $(docker images | grep risk-advisor | awk '{ print $3 }') >/dev/null 2>/dev/null

echo "Removing old simulator docker image"
docker rmi -f $(docker images | grep simulator | awk '{ print $3 }') >/dev/null 2>/dev/null

# stop when a command returns error
set -e

for name in "risk-advisor" "simulator"; do
	echo -e "\n\n################### $name ###################" | awk '{print toupper($0)}'

	echo "Building $name docker image"
	(cd $name; ./build.sh)

	echo "Tagging $name image"
	docker tag $(docker images | grep $name | awk '{ print $3 }' | head -1) "pposkrobko/$name:latest" >/dev/null

	echo "Pushing $name image"
	docker push "pposkrobko/$name:latest" >/dev/null
done
