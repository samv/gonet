package ethernet

//	"github.com/hsheth2/logs"

type Network_Writer struct {
	net *Network_Tap
}

func NewNetwork_Writer() (*Network_Writer, error) {
	return &Network_Writer{
		net: GlobalNetwork_Tap,
	}, nil
}

func (nw *Network_Writer) Write(data []byte, dst_mac *MAC_Address, ethertype EtherType) error {
	// build the ethernet header
	//	logs.Info.Println("Ethernet write request")
	index := getInternalIndex(dst_mac)
	//	logs.Info.Println("Found internal index")
	src_mac, err := globalSource_MAC_Table.search(index)
	if err != nil {
		return err
	}

	packet := make([]byte, ETH_HEADER_SZ+len(data))

	//	logs.Info.Println("Finished ARP lookup stuff")
	copy(packet, dst_mac.Data[:ETH_MAC_ADDR_SZ])
	copy(packet[ETH_MAC_ADDR_SZ:], src_mac.Data[:ETH_MAC_ADDR_SZ])
	packet[2*ETH_MAC_ADDR_SZ] = byte(ethertype >> 8)
	packet[2*ETH_MAC_ADDR_SZ+1] = byte(ethertype)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	copy(packet[ETH_HEADER_SZ:], data)

	// send packet
	//	logs.Info.Println("Send ethernet packet")
	if index == loopback_internal_index {
		nw.net.readBuf <- packet // TODO verify the packet is correctly built
		return nil
	} else {
		// logs.Info.Println("network_writer:", "write: full packet with ethernet header:", newPacket)
		return nw.net.write(packet)
	}
}

func (nw *Network_Writer) Close() error {
	return nil
}
