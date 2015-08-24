package ethernet

import (
	"errors"

	"github.com/hsheth2/logs"
)

type Ethernet_Header struct {
	//Rmac, Lmac *MAC_Address
	Packet []byte
}

var GlobalNetworkReader = func() *Network_Reader {
	x, err := NewNetwork_Reader()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}()

type Network_Reader struct {
	net       *Network_Tap
	proto_buf map[EtherType](chan *Ethernet_Header)
}

func NewNetwork_Reader() (*Network_Reader, error) {
	nr := &Network_Reader{
		net:       GlobalNetwork_Tap,
		proto_buf: make(map[EtherType](chan *Ethernet_Header)),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *Network_Reader) readAll() { // TODO terminate (using notifiers)
	for {
		data, err := nr.readFrame()
		// logs.Info.Println("Recv ethernet packet")
		if err != nil {
			// logs.Error.Println("ReadFrame failed:", err)
			continue
		}
		// logs.Trace.Println("network_reader readAll readFrame success")

		eth_protocol := EtherType(uint16(data[12])<<8 | uint16(data[13]))
		// logs.Trace.Println("Eth frame with protocol:", eth_protocol)
		if c, ok := nr.proto_buf[eth_protocol]; ok {
			// logs.Trace.Println("Something binded to protocol:", eth_protocol)
			// logs.Info.Println("Found that ethernet protocol is registered")

			ethHead := &Ethernet_Header{
				//Rmac: &MAC_Address{Data: data[ETH_MAC_ADDR_SZ : 2*ETH_MAC_ADDR_SZ]},
				//Lmac:   &MAC_Address{Data: data[0:ETH_MAC_ADDR_SZ]},
				Packet: data[ETH_HEADER_SZ:],
			}
			//			logs.Info.Println("Beginning to forward ethernet packet")
			select {
			case c <- ethHead:
				// logs.Info.Println("Forwarding ethernet packet")
			default:
				logs.Warn.Println("Dropping Ethernet packet: buffer full")
			}
		} else {
			logs.Warn.Println("Dropping Ethernet packet for wrong protocol:", eth_protocol)
		}
	}
}

func (nr *Network_Reader) Bind(proto EtherType) (chan *Ethernet_Header, error) {
	if _, exists := nr.proto_buf[proto]; exists {
		return nil, errors.New("Protocol already registered")
	} else {
		c := make(chan *Ethernet_Header, ETH_PROTOCOL_BUF_SZ)
		nr.proto_buf[proto] = c
		return c, nil
	}
}

func (nr *Network_Reader) Unbind(proto EtherType) error {
	// TODO write the unbind ether proto function
	return nil
}

func (nr *Network_Reader) readFrame() ([]byte, error) {
	return nr.net.read()
}
