package udp

import (
	"network/ipv4"
)

type reader struct {
	bytes     <-chan []byte
	port      Port // ports
	ipAddress *ipv4.Address
}

func NewReader(port Port, ip *ipv4.Address) (Reader, error) {
	bts, err := globalReadManager.bind(port, ip)
	if err != nil {
		return nil, err
	}

	return &reader{port: port, bytes: bts, ipAddress: ip}, nil
}

func (c *reader) Read(size int) ([]byte, error) {
	data := <-c.bytes
	if len(data) > size {
		data = data[:size]
	}
	return data, nil
}

func (c *reader) Close() error {
	return globalReadManager.unbind(c.port, c.ipAddress)
}
