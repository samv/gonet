package ethernet

import (
	"errors"

	"github.com/hsheth2/gonet/physical"

	"github.com/hsheth2/logs"
)

func initNetworkReadManager() {
	protoBufs = make(map[EtherType]*ethernetReader)
	go readAll()
}

var protoBufs map[EtherType](*ethernetReader) // TODO protect with sync.Mutex

func readAll() { // TODO terminate (using notifiers)
	for {
		data, err := readFrame()
		// /*logs*/logs.Info.Println("Recv ethernet packet")
		if err != nil {
			// logs.Error.Println("ReadFrame failed:", err)
			continue
		}
		// /*logs*/logs.Trace.Println("network_reader readAll readFrame success")

		ethProto := EtherType(uint16(data[12])<<8 | uint16(data[13]))
		// /*logs*/logs.Trace.Println("Eth frame with protocol:", eth_protocol)
		if c, ok := protoBufs[ethProto]; ok {
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

// Bind allows clients to bind to specific EtherTypes
func Bind(proto EtherType) (Reader, error) {
	if _, exists := protoBufs[proto]; exists {
		return nil, errors.New("Protocol already registered")
	}

	c, err := newEthernetReader(proto)
	if err != nil {
		return nil, err
	}
	protoBufs[proto] = c
	return c, nil
}

// Unbind allows clients to unbind from specific EtherTypes
func Unbind(proto EtherType) error {
	// TODO write the unbind ethernet proto function
	return nil
}

func readFrame() (d []byte, err error) {
	d, _, err = physical.Read()
	return
}
