#!/usr/bin/env bash

# should be run in gonet source directory
docker build -t gonet-linux .
docker run -it --privileged --rm -v `pwd`:/go/src/github.com/hsheth2/gonet gonet-linux:latest /bin/bash -c "cd /go/src/github.com/hsheth2/gonet && make && make test"
