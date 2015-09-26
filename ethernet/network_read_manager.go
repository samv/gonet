package ethernet

import (
	"errors"
	"network/physical"

	"github.com/hsheth2/logs"
)

var GlobalNetworkReadManager *NetworkReadManager

func initNetworkReadManager() *NetworkReadManager {
	x, err := newNetworkReadManager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}

type NetworkReadManager struct {
	protoBufs map[EtherType](*ethernetReader)
}

func newNetworkReadManager() (*NetworkReadManager, error) {
	nr := &NetworkReadManager{
		protoBufs: make(map[EtherType]*ethernetReader),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *NetworkReadManager) readAll() { // TODO terminate (using notifiers)
	for {
		data, err := nr.readFrame()
		// //ch logs.Info.Println("Recv ethernet packet")
		if err != nil {
			// logs.Error.Println("ReadFrame failed:", err)
			continue
		}
		// //ch logs.Trace.Println("network_reader readAll readFrame success")

		ethProto := EtherType(uint16(data[12])<<8 | uint16(data[13]))
		// //ch logs.Trace.Println("Eth frame with protocol:", eth_protocol)
		if c, ok := nr.protoBufs[ethProto]; ok {
			//logs.Trace.Println("Something binded to protocol:", eth_protocol)
			//logs.Info.Println("Found that ethernet protocol is registered")

			select {
			case c.input <- data:
			//logs.Trace.Println("Ethernet Data begin forwarded:", data)
			default:
				logs.Warn.Printf("Dropping Ethernet packet %v no space in input buffer\n", ethProto)
			}
		} else {
			logs.Warn.Println("Dropping Ethernet packet for wrong protocol:", ethProto)
		}
	}
}

func (nr *NetworkReadManager) Bind(proto EtherType) (Reader, error) {
	if _, exists := nr.protoBufs[proto]; exists {
		return nil, errors.New("Protocol already registered")
	}

	c, err := newEthernetReader(proto)
	if err != nil {
		return nil, err
	}
	nr.protoBufs[proto] = c
	return c, nil
}

func (nr *NetworkReadManager) Unbind(proto EtherType) error {
	// TODO write the unbind ethernet proto function
	return nil
}

func (nr *NetworkReadManager) readFrame() (d []byte, err error) {
	d, _, err = physical.Read()
	return
}
