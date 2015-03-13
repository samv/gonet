package main

import (
//	"errors"
	"fmt"
//	"net"
	"syscall"
	//"golang.org/x/net/ipv4"
)

type Network_Reader struct {
	fd int
}

const MAX_IP_PACKET_LEN = 65535

func NewNetwork_Reader() (*Network_Reader, error) {
	// 768 = htons(ETH_P_ALL) = htons(3)
    // see http://ideone.com/2eunQu
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, 768)

	if err != nil {
		fmt.Println("AF_PACKET socket connection")
		return nil, err
	}

	return &Network_Reader{
		fd: fd,
	}, nil
}

func (nr *Network_Reader) getNextPacket(buf []byte) (int, error) {
	return syscall.Read(nr.fd, buf)
}