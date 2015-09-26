package ethernet

import (
	"errors"
	"network/physical"

	"github.com/hsheth2/logs"
)

var GlobalNetworkReadManager = func() *Network_Read_Manager {
	x, err := newNetwork_Read_Manager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}()

type Network_Read_Manager struct {
	proto_buf map[EtherType](*ethernet_reader)
}

func newNetwork_Read_Manager() (*Network_Read_Manager, error) {
	nr := &Network_Read_Manager{
		proto_buf: make(map[EtherType]*ethernet_reader),
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
			//logs.Trace.Println("Something binded to protocol:", eth_protocol)
			//logs.Info.Println("Found that ethernet protocol is registered")

			select {
			case c.input <- data:
			//logs.Trace.Println("Ethernet Data begin forwarded:", data)
			default:
				logs.Warn.Printf("Dropping Ethernet packet %v no space in input buffer\n", eth_protocol)
			}
		} else {
			logs.Warn.Println("Dropping Ethernet packet for wrong protocol:", eth_protocol)
		}
	}
}

func (nr *Network_Read_Manager) Bind(proto EtherType) (Ethernet_Reader, error) {
	if _, exists := nr.proto_buf[proto]; exists {
		return nil, errors.New("Protocol already registered")
	} else {
		c, err := new_ethernet_reader(proto)
		if err != nil {
			return nil, err
		}
		nr.proto_buf[proto] = c
		return c, nil
	}
}

func (nr *Network_Read_Manager) Unbind(proto EtherType) error {
	// TODO write the unbind ethernet proto function
	return nil
}

func (nr *Network_Read_Manager) readFrame() (d []byte, err error) {
	d, _, err = physical.Read()
	return
}
