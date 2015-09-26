package ethernet

import "io"

type Ethernet_Reader interface {
	Read() (*Ethernet_Header, error)
	io.Closer
}

type Ethernet_Writer interface {
	Write(data []byte) (n int, err error)
	io.Closer
}
