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
	index := getInternalIndex(dst_mac)
	src_mac, err := globalSource_MAC_Table.search(index)
	if err != nil {
		return err
	}
	etherHead := append(append(
		dst_mac.Data[:ETH_MAC_ADDR_SZ],     // dst MAC
		src_mac.Data[:ETH_MAC_ADDR_SZ]...), // src MAC
		byte(ethertype>>8), byte(ethertype), // EtherType
	)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	newPacket := append(etherHead, data...)

	// send packet
	if index == loopback_internal_index {
		nw.net.readBuf <- newPacket // TODO verify the packet is correctly built
		return nil
	} else {
		//		logs.Info.Println("network_writer:", "write: full packet with ethernet header:", newPacket)
		return nw.net.write(newPacket)
	}
}

func (nw *Network_Writer) Close() error {
	return nil
}
