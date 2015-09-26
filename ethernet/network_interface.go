package ethernet

import "io"

func init() {
	GlobalNetworkReadManager = initNetworkReadManager()
	globalSourceMACTable = initSourceTable()
}

// Reader allows for reading packets from all interfaces for a specific EtherType
type Reader interface {
	Read() (*FrameHeader, error)
	io.Closer
}

// Writer allows for writing packets to a specific MAC address and EtherType
type Writer interface {
	Write(data []byte) (n int, err error)
	io.Closer
}
