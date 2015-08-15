package ethernet

import (
	"errors"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/water"
)

const TAP_NAME = "tap0"
const RX_QUEUE_SIZE = 2040 // TODO

type Network_Tap struct {
	ifce    *water.Interface
	readBuf chan []byte
}

var GlobalNetwork_Tap *Network_Tap = func() *Network_Tap {
	ifce, err := newNetwork_Tap(TAP_NAME)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	return ifce
}()

func newNetwork_Tap(ifname string) (*Network_Tap, error) {
	ifce, err := water.NewTAP(ifname)
	if err != nil {
		return nil, err
	}

	return &Network_Tap{
		ifce:    ifce,
		readBuf: make(chan []byte, RX_QUEUE_SIZE),
	}, nil
}

func (ntap *Network_Tap) write(data []byte) error {
	n, err := ntap.ifce.Write(data)
	if err != nil {
		return err
	}
	if len(data) != n {
		return errors.New("ifce failed to write all data")
	}
	return nil
}

func (ntap *Network_Tap) readAll() {
	for {
		rx, err := ntap.readOnce()
		if err != nil {
			continue // drop packet
		}
		select {
		case ntap.readBuf <- rx:
			// rx packet
		default:
			continue // drop packet
		}
	}
}

func (ntap *Network_Tap) readOnce() ([]byte, error) {
	buf := make([]byte, MAX_ETHERNET_FRAME_SZ)
	ln, err := ntap.ifce.Read(buf)
	if err != nil {
		return nil, err
	}
	logs.Trace.Println("readOnce:", buf)
	return buf[:ln], nil
}

func (ntap *Network_Tap) read() ([]byte, error) {
	return <-ntap.readBuf, nil
}
