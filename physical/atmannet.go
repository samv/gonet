package physical

import (
	"bytes"
	"fmt"
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
	a := &anetIO{
		ifce:    atmannet.DefaultDevice,
		rxBuf:   make([][]byte, 0),
		packets: make(chan []byte, 42),
	}
	go a.readAll()
	return a
}

func (anet *anetIO) Write(data []byte) (n int, err error) {
	return len(data), nil
}

func (anet *anetIO) readAll() {
	dev := anet.ifce

	fmt.Println("AtmanNet ready to read on:")
	fmt.Printf("  Mac address: %s\n", dev.MacAddr)
	fmt.Printf("  IP address:  %s\n", dev.IPAddr)

	for {
		dev.EventChannel.Wait()

		fmt.Println("AtmanNet woke up")

		for dev.Rx.CheckForResponses() {
			rsp := (*atmannet.NetifRxResponse)(dev.Rx.NextResponse())

			if rsp.Status > NETIF_RSP_NULL {
				dataLen := uint16(rsp.Status)
				dataBuf := dev.RxBuffers[rsp.ID]
				// efficient allocation is left as an exercise to the reader
				packet := dataBuf.Page.Data[rsp.Offset : rsp.Offset+dataLen]
				packetCopy := make([]byte, dataLen)
				copy(packetCopy, packet)

				fmt.Printf("Read %d byte(s) from atmannet\n", dataLen)
				anet.packets <- packetCopy
			} else {
				fmt.Printf("response: %+v\n", rsp)
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
