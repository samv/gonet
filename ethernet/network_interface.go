package ethernet

type Ethernet_Reader interface {
	Read() (*Ethernet_Header, error)
	Close() error
}

type Ethernet_Writer interface {
	Write(data []byte) (n int, err error)
	Close() error
}
