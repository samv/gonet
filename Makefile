# Makefile for Golang Network Stack

pkgs = network/ethernet network/arp network/ipv4/arpv4 network/ipv4/ipv4tps network/ipv4/ipv4src network/ipv4 network/udp network/tcp network/icmp network/ping

install: clean setup depend build
depend:
	go get -u github.com/hsheth2/logs
	go get -u github.com/hsheth2/notifiers
	go get -u github.com/hsheth2/water
	go get -u github.com/hsheth2/water/waterutil
build:
	go clean ${pkgs}
	go install ${pkgs}
clean:
	-rm ipv4/arpv4/ips.static.orig
	-rm ipv4/arpv4/ips_mac.static.orig
	-rm ipv4/arpv4/ips_mac.static
	-rm ipv4/ips.static
	-rm ethernet/mac.static.orig
	-rm ethernet/mac.static
	-rm ethernet/external_mac.static
	-rm *.test
	-rm *.pprof
	go clean ${pkgs}
setup:
	-./tap_setup.sh
	-./arp_setup.sh


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
test_udp: iptables
	./run_test.sh network/udp
test_tcp: iptables
	./run_test.sh network/tcp
test_ping:
	./run_test.sh network/ping
test_ethernet:
	# for testing water
	./run_test.sh network/ethernet


iptables:
	sudo iptables -I INPUT -p tcp --sport 20102 -j DROP
	sudo iptables -I INPUT -p tcp --dport 20102 -j DROP
	sudo iptables -I INPUT -p tcp --dport 20101 -j DROP
