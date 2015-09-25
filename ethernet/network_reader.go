package ethernet

import (
	"errors"

	"github.com/hsheth2/logs"
	"network/physical"
)

type Ethernet_Header struct {
	//Rmac, Lmac *MAC_Address
	Packet []byte
}

var GlobalNetworkReadManager = func() *Network_Read_Manager {
	x, err := newNetwork_Read_Manager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}()

type Network_Read_Manager struct {
	proto_buf map[EtherType](chan *Ethernet_Header)
}

func newNetwork_Read_Manager() (*Network_Read_Manager, error) {
	nr := &Network_Read_Manager{
		net:       physical.GlobalTapIO,
		proto_buf: make(map[EtherType](chan *Ethernet_Header)),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *Network_Read_Manager) readAll() { // TODO terminate (using notifiers)
	for {
		data, err := nr.readFrame()
		// //ch logs.Info.Println("Recv ethernet packet")
		if err != nil {
			// logs.Error.Println("ReadFrame failed:", err)
			continue
		}
		// //ch logs.Trace.Println("network_reader readAll readFrame success")

		eth_protocol := EtherType(uint16(data[12])<<8 | uint16(data[13]))
		// //ch logs.Trace.Println("Eth frame with protocol:", eth_protocol)
		if c, ok := nr.proto_buf[eth_protocol]; ok {
			// //ch logs.Trace.Println("Something binded to protocol:", eth_protocol)
			// //ch logs.Info.Println("Found that ethernet protocol is registered")

			ethHead := &Ethernet_Header{
				//Rmac: &MAC_Address{Data: data[ETH_MAC_ADDR_SZ : 2*ETH_MAC_ADDR_SZ]},
				//Lmac:   &MAC_Address{Data: data[0:ETH_MAC_ADDR_SZ]},
				Packet: data[ETH_HEADER_SZ:],
			}
			//			//ch logs.Info.Println("Beginning to forward ethernet packet")
			select {
			case c <- ethHead:
				// //ch logs.Info.Println("Forwarding ethernet packet")
			default:
				logs.Warn.Println("Dropping Ethernet packet: buffer full")
			}
		} else {
			logs.Warn.Println("Dropping Ethernet packet for wrong protocol:", eth_protocol)
		}
	}
}

func (nr *Network_Read_Manager) Bind(proto EtherType) (chan *Ethernet_Header, error) {
	if _, exists := nr.proto_buf[proto]; exists {
		return nil, errors.New("Protocol already registered")
	} else {
		c := make(chan *Ethernet_Header, ETH_PROTOCOL_BUF_SZ)
		nr.proto_buf[proto] = c
		return c, nil
	}
}

func (nr *Network_Read_Manager) Unbind(proto EtherType) error {
	// TODO write the unbind ether proto function
	return nil
}

func (nr *Network_Read_Manager) readFrame() (d []byte, err error) {
	_, d, err = physical.Physical_IO.Read()
	return
}
