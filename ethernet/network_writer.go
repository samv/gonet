package ethernet

import (
	"network/physical"
	//"github.com/hsheth2/logs"
)

type ethernetWriter struct {
	dstMAC, srcMAC *MACAddress
	ethertype      EtherType
	index          physical.InternalIndex
}

// NewEthernetWriter allows for the writing to a given MAC Address and EtherType
func NewEthernetWriter(dstMAC *MACAddress, ethertype EtherType) (Writer, error) {
	index := getInternalIndex(dstMAC)
	//	/*logs*/logs.Info.Println("Found internal index")
	srcMAC, err := globalSourceMACTable.search(index)
	if err != nil {
		return nil, err
	}

	return &ethernetWriter{
		dstMAC:    dstMAC,
		srcMAC:    srcMAC,
		ethertype: ethertype,
		index:     index,
	}, nil
}

// Write is a blocking write call
func (nw *ethernetWriter) Write(data []byte) (int, error) {
	// build the ethernet header
	//	/*logs*/logs.Info.Println("Ethernet write request")
	packet := make([]byte, ethHeaderSize+len(data))

	//	/*logs*/logs.Info.Println("Finished ARP lookup stuff")
	copy(packet, nw.dstMAC.Data[:ethMACAddressSize])
	copy(packet[ethMACAddressSize:], nw.srcMAC.Data[:ethMACAddressSize])
	packet[2*ethMACAddressSize] = byte(nw.ethertype >> 8)
	packet[2*ethMACAddressSize+1] = byte(nw.ethertype)
	//fmt.Println("My header:", etherHead)

	// add on the ethernet header
	copy(packet[ethHeaderSize:], data)

	// send packet
	//logs.Trace.Println("Ethernet sending packet:", packet)
	return physical.Write(nw.index, packet)
}

func (nw *ethernetWriter) Close() error {
	return nil
}

// WriteSingle is a helper method that is used for one-time sends that do not require a full Writer
func WriteSingle(dstMAC *MACAddress, ethertype EtherType, data []byte) (int, error) {
	nw, err := NewEthernetWriter(dstMAC, ethertype)
	if err != nil {
		return 0, err
	}
	defer nw.Close()

	return nw.Write(data)
}
