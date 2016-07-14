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
		if c, ok := protoBufs[ethProto]; ok {
			select {
			case c.input <- data:
				logs.Trace.Printf("EtherType is %.4x (payload %d byte(s))", ethProto, len(data)-14)
				//
			default:
				logs.Warn.Printf("Dropping Ethernet packet %v no space in input buffer\n", ethProto)
			}
		} else {
			logs.Warn.Printf("ignoring ethernet frame with EtherType: %.4x", ethProto)
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
	_, exists := protoBufs[proto]
	if exists {
		delete(protoBufs, proto)
		return nil
	}
	return errors.New("That protocol is not bound!")
}

func readFrame() (d []byte, err error) {
	d, _, err = physical.Read()
	return
}
