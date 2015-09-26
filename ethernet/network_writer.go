package ethernet

import (
	"network/physical"
	//"github.com/hsheth2/logs"
)

type ethernetWriter struct {
	dstMAC, srcMAC *MACAddress
	ethertype        EtherType
	index            physical.InternalIndex
}

func NewEthernetWriter(dstMAC *MACAddress, ethertype EtherType) (Writer, error) {
	index := getInternalIndex(dstMAC)
	//	//ch logs.Info.Println("Found internal index")
	srcMAC, err := globalSourceMACTable.search(index)
	if err != nil {
		return nil, err
	}

	return &ethernetWriter{
		dstMAC:   dstMAC,
		srcMAC:   srcMAC,
		ethertype: ethertype,
		index:     index,
	}, nil
}

// blocking write call
func (nw *ethernetWriter) Write(data []byte) (int, error) {
	// build the ethernet header
	//	//ch logs.Info.Println("Ethernet write request")
	packet := make([]byte, ethHeaderSize+len(data))

	//	//ch logs.Info.Println("Finished ARP lookup stuff")
	copy(packet, nw.dstMAC.Data[:ethMACAddressSize])
	copy(packet[ethMACAddressSize:], nw.srcMAC.Data[:ethMACAddressSize])
	packet[2*ethMACAddressSize] = byte(nw.ethertype >> 8)
	packet[2*ethMACAddressSize+1] = byte(nw.ethertype)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	copy(packet[ethHeaderSize:], data)

	// send packet
	//logs.Trace.Println("Ethernet sending packet:", packet)
	return physical.Write(nw.index, packet) // TODO do not use directly?
}

func (nw *ethernetWriter) Close() error {
	return nil
}

// helper method for one-time sends
func WriteSingle(dstMAC *MACAddress, ethertype EtherType, data []byte) (int, error) {
	nw, err := NewEthernetWriter(dstMAC, ethertype)
	if err != nil {
		return 0, err
	}
	defer nw.Close()

	return nw.Write(data)
}
