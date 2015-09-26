package ethernet

import (
	"github.com/hsheth2/logs"
)

type Ethernet_Header struct {
	//Rmac, Lmac *MAC_Address
	Packet []byte
}

type ethernet_reader struct {
	ethertype EtherType
	input     chan []byte
	processed chan *Ethernet_Header
}

func new_ethernet_reader(etp EtherType) (*ethernet_reader, error) {
	ethr := &ethernet_reader{
		ethertype: etp,
		input:     make(chan []byte, ETH_PROTOCOL_BUF_SZ),
		processed: make(chan *Ethernet_Header, ETH_PROTOCOL_BUF_SZ),
	}

	go ethr.readAll()

	return ethr, nil
}

func (ethr *ethernet_reader) readAll() {
	for {
		//logs.Trace.Println("Ethernet reader attempting to get work")
		data := <-ethr.input
		//logs.Trace.Println("Ethernet reader recieved packet")

		ethHead := &Ethernet_Header{
			//Rmac: &MAC_Address{Data: data[ETH_MAC_ADDR_SZ : 2*ETH_MAC_ADDR_SZ]},
			//Lmac:   &MAC_Address{Data: data[0:ETH_MAC_ADDR_SZ]},
			Packet: data[ETH_HEADER_SZ:],
		}
		//			//ch logs.Info.Println("Beginning to forward ethernet packet")
		select {
		case ethr.processed <- ethHead:
			//logs.Trace.Println("Forwarding ethernet packet")
		default:
			logs.Warn.Println("Dropping Ethernet packet: buffer full")
		}
	}
}

// blocking read call
func (ethr *ethernet_reader) Read() (*Ethernet_Header, error) {
	return <-ethr.processed, nil
}

func (ethr *ethernet_reader) Close() error {
	// TODO unbind
	// TODO close input channel
	// TODO close output channel
	return nil
}
