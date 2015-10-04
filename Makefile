# Makefile for Golang Network Stack

SHELL = /bin/bash

install: clean setup depend build
depend:
	go get -u github.com/hsheth2/logs
	go get -u github.com/hsheth2/notifiers
	go get -u github.com/pkg/profile
	go get -u github.com/hsheth2/water
	go get -u github.com/hsheth2/water/waterutil
build:
	go clean ./...
	go install ./...
clean:
	-rm -rf *.static.orig
	-rm -rf *.static
	-rm -f *.test
	-rm -f *.cover
	-rm -f *.html
	go clean ./...
setup:
	-./tap_setup.sh
	-./arp_setup.sh

# line counting
lines:
	find ./ -name '*.go' | xargs wc -l

# Error Checking
vet:
	go vet ./...
fmt:
	@echo "Formatting Files..."
	goimports -l -w ./
	@echo "Finished Formatting"
lint:
	golint ./...


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
