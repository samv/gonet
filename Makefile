# Makefile for Golang Network Stack

SHELL = /bin/bash

pkgs = network/physical network/ethernet network/arp network/ipv4/arpv4 network/ipv4/ipv4tps network/ipv4/ipv4src network/ipv4 network/udp network/tcp network/icmp network/ping

install: clean setup depend build
depend:
	go get -u github.com/hsheth2/logs
	go get -u github.com/hsheth2/notifiers
	go get -u github.com/pkg/profile
	go get -u github.com/hsheth2/water
	go get -u github.com/hsheth2/water/waterutil
build:
	go clean ${pkgs}
	go install ${pkgs}
clean:
	-rm -rf *.static.orig
	-rm -rf *.static
	-rm -f *.test
	-rm -f *.cover
	-rm -f *.html
	-rm runStack scaleTest local_latency
	go clean ${pkgs}
setup:
	-./tap_setup.sh
	-./arp_setup.sh

# line counting
lines_all:
	find ./ -name '*.go' -o -name '*.py' -o -name '*.c' -o -name '*.sh' | xargs wc -l
lines_go:
	find ./ -name '*.go' | xargs wc -l

# Error Checking
vet:
	go vet ${pkgs}
fmt:
	./auto-format.sh
	# go fmt ${pkgs}


# Different tests that could be run on the network's code
test: test_others test_network
test_others:
	./run_test.sh github.com/hsheth2/logs
	./run_test.sh github.com/hsheth2/notifiers
test_network: test_udp test_tcp test_ping
test_udp:
	./run_test.sh network/udp
test_tcp:
	./run_test.sh network/tcp
test_ping:
	./run_test.sh network/ping
test_tap:
	# for testing water
	./run_test.sh network/ethernet

# Performance
#latency:
#	sh latency_test.sh 10
#local_latency:
#	-sudo pkill local_latency
#	-sudo pkill local_latency
#	go build local_latency.go
#	sudo setcap CAP_NET_RAW=epi ./local_latency
#	time (./local_latency)
throughput:
	bash throughput_test.sh
#scale:
#	-sudo pkill scaleTest
#	-sudo pkill tapip
#	go build scaleTest.go
#	sudo setcap CAP_NET_RAW=epi ./scaleTest
#	./scaleTest > /dev/null 2>&1 &
#	sleep 1
#	python tcp/tcp_client.py
#	pkill scaleTest
