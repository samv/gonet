# Makefile for go program
# iptables command: sudo iptables -I INPUT -p tcp --sport 20102 -j DROP

pkgs = network/logs network/notifiers network/etherp network/ipv4p network/udpp network/tcpp

install:
	go install ${pkgs}
vet:
	go vet ${pkgs}
fmt:
	go fmt ${pkgs}
test:
	go test network/udpp