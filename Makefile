# Makefile for Golang Network Stack

pkgs = network/etherp network/ipv4p network/udpp network/tcpp network/icmpp

install:
	go get github.com/hsheth2/logs
	go get github.com/hsheth2/notifiers
	go clean ${pkgs}
	go install ${pkgs}
vet:
	go vet ${pkgs}
fmt:
	go fmt ${pkgs}


# Different tests that could be run on the network's code
test: test_others test_network
test_others:
	./run_test.sh github.com/hsheth2/logs
	./run_test.sh github.com/hsheth2/notifiers
test_network: test_udp test_tcp test_icmp
test_udp: iptables
	./run_test.sh network/udpp
test_tcp: iptables
	./run_test.sh network/tcpp
test_icmp:
	./run_test.sh network/icmpp

iptables:
	sudo iptables -I INPUT -p tcp --sport 20102 -j DROP
	sudo iptables -I INPUT -p tcp --dport 20102 -j DROP
	sudo iptables -I INPUT -p tcp --dport 20101 -j DROP
