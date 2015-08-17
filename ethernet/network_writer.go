package ethernet

import (
	"bytes"

	"github.com/hsheth2/logs"
)

type Network_Writer struct {
	net *Network_Tap
}

func NewNetwork_Writer() (*Network_Writer, error) {
	return &Network_Writer{
		net: GlobalNetwork_Tap,
	}, nil
}

var loopback_mac_address *MAC_Address = &MAC_Address{Data: []byte{0, 0, 0, 0, 0, 0}}

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

	// send packet
	if bytes.Equal(src_mac.Data, loopback_mac_address.Data) { // TODO find a better, dynamic way to do this
		nw.net.readBuf <- newPacket // TODO verify the packet is correctly built
		return nil
	} else {
		logs.Info.Println("network_writer:", "write: full packet with ethernet header:", newPacket, "with ifindex:", addr.IF_index)
		return nw.net.write(newPacket)
	}
}

func (nw *Network_Writer) Close() error {
	return nil
}
