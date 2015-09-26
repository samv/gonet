package ethernet

import (
	"network/physical"
	//"github.com/hsheth2/logs"
)

type ethernet_writer struct {
	dst_mac, src_mac *MAC_Address
	ethertype        EtherType
	index            physical.InternalIndex
}

func NewEthernet_Writer(dst_mac *MAC_Address, ethertype EtherType) (Ethernet_Writer, error) {
	index := getInternalIndex(dst_mac)
	//	//ch logs.Info.Println("Found internal index")
	src_mac, err := globalSource_MAC_Table.search(index)
	if err != nil {
		return nil, err
	}

	return &ethernet_writer{
		dst_mac:   dst_mac,
		src_mac:   src_mac,
		ethertype: ethertype,
		index:     index,
	}, nil
}

// blocking write call
func (nw *ethernet_writer) Write(data []byte) (int, error) {
	// build the ethernet header
	//	//ch logs.Info.Println("Ethernet write request")
	packet := make([]byte, ETH_HEADER_SZ+len(data))

	//	//ch logs.Info.Println("Finished ARP lookup stuff")
	copy(packet, nw.dst_mac.Data[:ETH_MAC_ADDR_SZ])
	copy(packet[ETH_MAC_ADDR_SZ:], nw.src_mac.Data[:ETH_MAC_ADDR_SZ])
	packet[2*ETH_MAC_ADDR_SZ] = byte(nw.ethertype >> 8)
	packet[2*ETH_MAC_ADDR_SZ+1] = byte(nw.ethertype)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	copy(packet[ETH_HEADER_SZ:], data)

	// send packet
	//logs.Trace.Println("Ethernet sending packet:", packet)
	return physical.Write(nw.index, packet) // TODO do not use directly?
}

func (nw *ethernet_writer) Close() error {
	return nil
}

// helper method for one-time sends
func EthernetWriteOne(dst_mac *MAC_Address, ethertype EtherType, data []byte) (int, error) {
	nw, err := NewEthernet_Writer(dst_mac, ethertype)
	if err != nil {
		return 0, err
	}

	return nw.Write(data)
}
