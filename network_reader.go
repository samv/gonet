package main

import (
	"errors"
	"fmt"
	"net"
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
	buffers map[string](map[uint8](chan []byte))
	//buffers map[uint8](chan []byte)
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
		fd: fd,
		//buffers: make(map[uint8](chan []byte)),
		buffers: make(map[string](map[uint8](chan []byte))),
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

		// TODO: assemble IP fragments

		protocol := uint8(buf[9])
		ip := net.IPv4(buf[12], buf[13], buf[14], buf[15]).String()

		if protoBuf, found := nr.buffers[ip]; found {
			if c, foundProto := protoBuf[protocol]; foundProto {
				go func() { c <- buf }()
			}
		}
	}
}

func (nr *Network_Reader) bind(ip string, protocol uint8) (<-chan []byte, error) {
	// TODO: implement forwarding
	_, ipOk := nr.buffers[ip]
	if !ipOk {
		nr.buffers[ip] = make(map[uint8](chan []byte))
	}
	protoBuf, _ := nr.buffers[ip]
	if _, ok := protoBuf[protocol]; !ok {
		// doesn't exist in map already
		protoBuf[protocol] = make(chan []byte, 1)

		ret, _ := protoBuf[protocol]
		return ret, nil
	}
	return nil, errors.New("Protocol already binded to.")
}

func (nr *Network_Reader) unbind(ip string, protocol uint8) error {
	// TODO: take into account the protocol

	protoBuf, ipOk := nr.buffers[ip]
	if !ipOk {
		return errors.New("IP not bound, cannot unbind")
	}

	if _, ok := protoBuf[protocol]; ok {
		delete(protoBuf, protocol)
		if len(protoBuf) == 0 {
			delete(nr.buffers, ip)
		}
		return nil
	}
	return errors.New("Not binded, can't unbind.")
}

func (nr *Network_Reader) getNextPacket(buf []byte) (int, error) {
	return syscall.Read(nr.fd, buf)
}
