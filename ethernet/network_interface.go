package ethernet

// TODO finish this

type network_close interface {
	Close() error
}

type Ethernet_Reader interface {
	network_close

}

type Ethernet_Writer interface {
	network_close
	Write(data []byte) (n int, err error)
}
