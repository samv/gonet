package physical

import (
	"errors"

	"github.com/hsheth2/logs"
)

type loopbackIO struct {
	readBuf chan []byte
}

var globalLoopbackIO *loopbackIO

func loInit() *loopbackIO {
	lo, err := newLoopbackIO()
	if err != nil {
		logs.Error.Fatalln(err)
	}
	return lo
}

func newLoopbackIO() (*loopbackIO, error) {
	lo := &loopbackIO{
		readBuf: make(chan []byte, rxQueueSize),
	}

	return lo, nil
}

func (lo *loopbackIO) getInput() chan []byte {
	return lo.readBuf
}

// blocking write to loopback "interface"
func (lo *loopbackIO) Write(data []byte) (int, error) {
	lo.readBuf <- data
	///*logs*/logs.Info.Println("Finished loopback write")
	return len(data), nil
}

func (lo *loopbackIO) Read() ([]byte, error) {
	///*logs*/logs.Trace.Println("read packet off network_tap")
	res := <-lo.readBuf
	if len(res) == 0 {
		return nil, errors.New("Channel is closed!")
	}
	return <-lo.readBuf, nil
}

func (lo *loopbackIO) Close() error {
	close(lo.readBuf)
	return nil
}
