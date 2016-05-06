FROM golang:1.6

RUN apt-get update
RUN apt-get -y install sudo apt-utils net-tools iptables

#RUN useradd -m docker && echo "docker:docker" | chpasswd && adduser docker sudo
#USER docker

# VOLUME ["/go/src/github.com/hsheth2/gonet"]
# ADD . /go/src/github.com/hsheth2/gonet

CMD /bin/bash

