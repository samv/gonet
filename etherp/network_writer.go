package etherp

import (
	"errors"
	"syscall"
	//"github.com/hsheth2/logs"
)

type Network_Writer struct {
	fd       int
	sockAddr syscall.Sockaddr
}

func NewNetwork_Writer() (*Network_Writer, error) {
	fd, err := syscall.Socket(AF_PACKET, SOCK_RAW, HTONS_ETH_P_ALL)
	if err != nil {
		return nil, errors.New("Write's socket failed")
	}

	addr := getSockAddr()

	/*err = syscall.Sendto(fd, []byte{0x08, 0x00, 0x27, 0x9e, 0x29, 0x63, 0x08, 0x00, 0x27, 0x9e, 0x29, 0x63, 0x08, 0x00}, 0, addr) //Random bytes
	  if err != nil {
	      fmt.Println("ERROR returned by syscall.Sendto", err)
	  } else {
	      fmt.Println("Sent the test packet")
	  }*/

	return &Network_Writer{
		fd:       fd,
		sockAddr: addr,
	}, nil
}

func (nw *Network_Writer) Write(data []byte) error {
	// build the ethernet header
	/*etherHead :=  append(append(
	    myMACSlice, // dst MAC
	    myMACSlice...), // src MAC
	    0x08, 0x00, // ethertype (IP)
	)*/
	// TODO: decide MAC address dynamically
	etherHead := []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0,
	}
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	newPacket := append(etherHead, data...)
	//fmt.Println("Full Packet with ethernet header:", newPacket)

	//logs.Trace.Println("Ethernet Writing:", newPacket)
	return syscall.Sendto(nw.fd, newPacket, 0, nw.sockAddr)
}

func (nw *Network_Writer) Close() error {
	return syscall.Close(nw.fd)
}
