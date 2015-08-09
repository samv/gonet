package ethernet

import (
	"errors"
	"syscall"
)

type Network_Writer struct {
	fd int
}

func NewNetwork_Writer() (*Network_Writer, error) {
	fd, err := syscall.Socket(AF_PACKET, SOCK_RAW, HTONS_ETH_P_ALL)
	if err != nil {
		return nil, errors.New("Write's socket failed")
	}

	return &Network_Writer{
		fd: fd,
	}, nil
}

func (nw *Network_Writer) Write(data []byte, addr *Ethernet_Addr, ethertype EtherType) error {
	// build the ethernet header
	src_mac, err := GlobalSource_MAC_Table.findByIf(addr.IF_index)
	if err != nil {
		return err
	}
	etherHead := append(append(
		addr.MAC.Data[:ETH_MAC_ADDR_SZ],    // dst MAC
		src_mac.Data[:ETH_MAC_ADDR_SZ]...), // src MAC
		byte(ethertype>>8), byte(ethertype), // EtherType
	)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	newPacket := append(etherHead, data...)
	//logs.Info.Println("Full Packet with ethernet header:", newPacket)

	// send packet
	return syscall.Sendto(nw.fd, newPacket, 0, getSockAddr(addr))
}

func (nw *Network_Writer) Close() error {
	return syscall.Close(nw.fd) // TODO notify upper layers of close
}
