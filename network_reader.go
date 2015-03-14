package main

import (
	"errors"
	"fmt"
	//	"net"
	"syscall"
	//"golang.org/x/net/ipv4"
)

/*type Network_Reader_IP struct {
    dst string
    protocol uint8
    buffer chan []byte
}

func NewNetwork_Reader_IP(dst string, protocol uint8) (*Network_Reader_IP, error) {

}*/

type Network_Reader struct {
	fd      int
	buffers map[uint8](chan []byte)
}

const MAX_IP_PACKET_LEN = 65535

func NewNetwork_Reader() (*Network_Reader, error) {
	// 768 = htons(ETH_P_ALL) = htons(3)
	// see http://ideone.com/2eunQu

	// 17 = AF_PACKET
	// see http://ideone.com/TGYlGc
	fd, err := syscall.Socket(17, syscall.SOCK_DGRAM, 768)

	if err != nil {
		fmt.Println("AF_PACKET socket connection")
		return nil, err
	}

	nr := &Network_Reader{
		fd:      fd,
		buffers: make(map[uint8](chan []byte)),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *Network_Reader) readAll() {
	for {
		buf := make([]byte, MAX_IP_PACKET_LEN)
		ln, err := nr.getNextPacket(buf)

		if err != nil {
			fmt.Println(err)
		}

		buf = buf[:ln]
		if len(buf) <= 20 {
			continue
		}

		protocol := uint8(buf[9])
		if c, found := nr.buffers[protocol]; found {
			go func() { c <- buf }()
		} else {
			//fmt.Println("Dropping packet from readAll; no matching connection for protocol: ", protocol)
			//fmt.Print("d")
		}
	}
}

func (nr *Network_Reader) bind(ip string, protocol uint8) (<-chan []byte, error) {
	// TODO: this does not take into account the ip requested... implement longest prefix match

	if _, ok := nr.buffers[protocol]; !ok {
		// doesn't exist in map already
		nr.buffers[protocol] = make(chan []byte, 1)
		return nr.buffers[protocol], nil
	}
	return nil, errors.New("Protocol already binded to.")
}

func (nr *Network_Reader) unbind(ip string, protocol uint8) error {
	// TODO: take into account the protocol

	if _, ok := nr.buffers[protocol]; ok {
		delete(nr.buffers, protocol)
		return nil
	}
	return errors.New("Not binded, can't unbind.")
}

func (nr *Network_Reader) getNextPacket(buf []byte) (int, error) {
	return syscall.Read(nr.fd, buf)
}
