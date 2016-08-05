#!/bin/bash

docker run -it --rm -w /go/src/github.com/Dataman-Cloud/rolex -v $(pwd):/go/src/github.com/Dataman-Cloud/rolex golang:1.5.4 make
curl https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz | tar xzv
docker-compose -p rolex -f deploy/docker-compose.yml stop
docker-compose -p rolex -f deploy/docker-compose.yml rm -f
docker-compose -p rolex -f deploy/docker-compose.yml up -d
