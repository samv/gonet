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

	nt := &Network_Tap{
		ifce:    ifce,
		readBuf: make(chan []byte, RX_QUEUE_SIZE),
	}

	go nt.readAll()

	return nt, nil
}

func (ntap *Network_Tap) write(data []byte) error {
	n, err := ntap.ifce.Write(data)
	if err != nil {
		return err
	}
	if len(data) != n {
		return errors.New("ifce failed to write all data")
	}
	//logs.Info.Println("Finished write")
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
			//logs.Trace.Println("Forwarded packet in readAll")
		default:
			logs.Warn.Println("Dropped packet in Network_Tap readAll")
			continue
		}
	}
}

func (ntap *Network_Tap) readOnce() ([]byte, error) {
	buf := make([]byte, MAX_ETHERNET_FRAME_SZ)
	ln, err := ntap.ifce.Read(buf)
	if err != nil {
		return nil, err
	}
	//logs.Trace.Println("readOnce:", buf[:ln])
	return buf[:ln], nil
}

func (ntap *Network_Tap) read() ([]byte, error) {
	//logs.Trace.Println("read packet off network_tap")
	return <-ntap.readBuf, nil
}

func (ntap *Network_Tap) close() error {
	return ntap.ifce.Close()
}
