#!/usr/bin/env bash

docker rm gonet

# should be run in gonet source directory
docker build -t gonet-linux .
docker run -it --privileged --name gonet -v `pwd`:/go/src/github.com/hsheth2/gonet gonet-linux:latest
