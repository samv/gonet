#!/usr/bin/env bash

docker rm gonet

# should be run in gonet source directory
docker pull hsheth2/gonet-base
docker run -it --privileged --name gonet -v `pwd`:/go/src/github.com/hsheth2/gonet hsheth2/gonet-base:latest
