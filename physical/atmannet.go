package physical

import (
	"bytes"
	"io"

	atmannet "atman/net"
)

type anetIO struct {
	ifce    *atmannet.Device
	rxBuf   [][]byte
	packets chan []byte
}

const (
	NETIF_RSP_NULL = 1
)

var globalAnetIO *anetIO

func anetInit() *anetIO {
	return &anetIO{
		ifce:    atmannet.DefaultDevice,
		rxBuf:   make([][]byte, 0),
		packets: make(chan []byte, 42),
	}
}

func (anet *anetIO) Write(data []byte) (n int, err error) {
	return len(data), nil
}

func (anet *anetIO) readAll() {
	dev := anet.ifce
	for {
		dev.EventChannel.Wait()

		for dev.Rx.CheckForResponses() {
			rsp := (*atmannet.NetifRxResponse)(dev.Rx.NextResponse())

			if rsp.Status > NETIF_RSP_NULL {
				r := newPacketReader(dev, rsp)

				// efficient allocation is left as an exercise to the reader
				var buf = make([]byte, 2048)
				n, _ := r.Read(buf)
				buf = buf[:n]
				anet.packets <- buf
			}

			enqueueRequest(dev, rsp.ID)
		}

		if notify := dev.Rx.PushRequests(); notify {
			dev.EventChannel.Notify()
		}
	}
}

func (anet *anetIO) dequeue() []byte {
	next := anet.rxBuf[0]
	anet.rxBuf = anet.rxBuf[1:] // see above comment
	return next
}

func (anet *anetIO) getInput() chan []byte {
	return anet.packets
}

func (anet *anetIO) Read() ([]byte, error) {
	return <-anet.packets, nil
}

func (anet *anetIO) Close() error {
	return nil
}

// enqueueRequest recycles the buffer for more packets
func enqueueRequest(dev *atmannet.Device, id uint16) {
	req := (*atmannet.NetifRxRequest)(dev.Rx.NextRequest())
	req.ID = id
	req.Gref = dev.RxBuffers[id].Gref
}

func newPacketReader(dev *atmannet.Device, rsp *atmannet.NetifRxResponse) io.Reader {
	var (
		len    = uint16(rsp.Status)
		buf    = dev.RxBuffers[rsp.ID]
		packet = buf.Page.Data[rsp.Offset : rsp.Offset+len]
	)

	return bytes.NewReader(packet)
}
