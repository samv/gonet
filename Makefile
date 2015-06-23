# Makefile for go program
# iptables command: sudo iptables -I INPUT -p tcp --sport 20102 -j DROP

pkgs = network/etherp network/ipv4p network/udpp network/tcpp

install:
	go get github.com/hsheth2/logs
	go get github.com/hsheth2/notifiers
	go install ${pkgs}
vet:
	go vet ${pkgs}
fmt:
	go fmt ${pkgs}
test:
	go test github.com/hsheth2/logs
    go test github.com/hsheth2/notifiers
	go test network/udpp
