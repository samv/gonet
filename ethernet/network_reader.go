package ethernet

import (
	"errors"

	"network/physical"

	"github.com/hsheth2/logs"
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
	_, d, err = physical.Physical_IO.Read()
	return
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
