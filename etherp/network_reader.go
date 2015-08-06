package etherp

import (
	//"errors"
	"github.com/hsheth2/logs"
	//"net"
	"bytes"
	"errors"
	"syscall"
)

var GlobalNetworkReader = func() *Network_Reader {
	x, err := NewNetwork_Reader()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}()

type Network_Reader struct {
	fd        int
	last      []byte
	proto_buf map[uint16](chan []byte)
}

func NewNetwork_Reader() (*Network_Reader, error) {
	fd, err := syscall.Socket(AF_PACKET, SOCK_RAW, HTONS_ETH_P_ALL)

	if err != nil {
		logs.Error.Println("AF_PACKET socket connection")
		return nil, err
	}

	nr := &Network_Reader{
		fd:        fd,
		last:      nil,
		proto_buf: make(map[uint16](chan []byte)),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *Network_Reader) readAll() {
	for {
		data, err := nr.readFrame()
		if err != nil {
			logs.Info.Println("ReadFrame failed:", err)
		}

		eth_protocol := uint16(data[12])<<8 | uint16(data[13])
		if c, ok := nr.proto_buf[eth_protocol]; ok {
			go func() { c <- data[ETH_HEADER_SZ:] }()
		} else {
			//logs.Info.Println("Dropping Ethernet packet for wrong protocol:", eth_protocol)
		}
	}
}

func (nr *Network_Reader) Bind(proto uint16) (chan []byte, error) {
	if _, exists := nr.proto_buf[proto]; exists {
		return nil, errors.New("Protocol already registered")
	} else {
		c := make(chan []byte, ETH_PROTOCOL_BUF_SZ)
		nr.proto_buf[proto] = c
		return c, nil
	}
}

func (nr *Network_Reader) Unbind(proto uint16) error {
	// TODO write the unbind ether proto function
	return nil
}

func (nr *Network_Reader) readFrame() ([]byte, error) {
	buf := make([]byte, MAX_ETHERNET_FRAME_SZ)
	// read twice to account for the double receiving TODO fix the double reading somehow
	//syscall.Read(nr.fd, buf)
	ln, err := syscall.Read(nr.fd, buf)
	if bytes.Equal(buf[:ln], nr.last) || err != nil {
		//logs.Info.Println("Dropping double read packet")
		return nr.readFrame()
	}
	nr.last = buf[:ln]
	return buf[:ln], err
}
