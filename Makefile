# Makefile for go program
# iptables command: sudo iptables -I INPUT -p tcp --sport 20102 -j DROP

pkgs = network/etherp network/ipv4p network/udpp network/tcpp

install:
	go get github.com/hsheth2/logs
	go get github.com/hsheth2/notifiers
	#sudo setcap CAP_NET_RAW=epi ./etherp/network_reader.go
	#sudo setcap CAP_NET_ADMIN=epi ./etherp/network_reader.go
	go install ${pkgs}
vet:
	go vet ${pkgs}
fmt:
	go fmt ${pkgs}
test:
	./run_test.sh github.com/hsheth2/logs
	./run_test.sh github.com/hsheth2/notifiers
	./run_test.sh network/udpp
