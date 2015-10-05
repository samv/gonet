package physical

import (
	"errors"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/water"
)

const tapName = "tap0"

type tapIO struct {
	ifce    *water.Interface
	readBuf chan []byte
}

var globalTapIO *tapIO

func tapInit() *tapIO {
	ifce, err := newTapIO(tapName)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	err = ifce.ifce.SetPersistent(true)
	if err != nil {
		logs.Warn.Println("Failed to make persistent:", err)
	}
	return ifce
}

func newTapIO(ifname string) (*tapIO, error) {
	ifce, err := water.NewTAP(ifname)
	if err != nil {
		return nil, err
	}

	tap := &tapIO{
		ifce:    ifce,
		readBuf: make(chan []byte, rxQueueSize),
	}

	go tap.readAll()

	return tap, nil
}

func (tap *tapIO) getInput() chan []byte {
	return tap.readBuf
}

func (tap *tapIO) Write(data []byte) (n int, err error) {
	n, err = tap.ifce.Write(data)
	if err != nil {
		return 0, err
	}
	if len(data) != n {
		return n, errors.New("ifce failed to write all data")
	}
	////ch logs.Info.Println("Finished write")
	return n, nil
}

func (tap *tapIO) readAll() {
	for {
		rx, err := tap.readOnce()
		if err != nil {
			continue // drop packet
		}
		select {
		case tap.readBuf <- rx:
			////ch logs.Trace.Println("Forwarded packet in readAll")
		default:
			logs.Warn.Println("Dropped packet in Network_Tap readAll")
			continue
		}
	}
}

func (tap *tapIO) readOnce() ([]byte, error) {
	buf := make([]byte, maxFrameSize)
	ln, err := tap.ifce.Read(buf)
	if err != nil {
		return nil, err
	}
	////ch logs.Trace.Println("readOnce:", buf[:ln])
	return buf[:ln], nil
}

func (tap *tapIO) Read() ([]byte, error) {
	////ch logs.Trace.Println("read packet off network_tap")
	return <-tap.readBuf, nil
}

func (tap *tapIO) Close() error {
	return tap.ifce.Close()
}
