package udp

import (
	"network/ipv4/ipv4tps"
)

const MAX_UDP_PACKET_LEN = 65507
const UDP_RCV_BUF_SZ = 1000

type UDP_Reader struct {
	manager   *UDP_Read_Manager
	bytes     <-chan []byte
	port      uint16 // ports
	ipAddress *ipv4tps.IPaddress
}

func NewUDP(x *UDP_Read_Manager, port uint16, ip *ipv4tps.IPaddress) (*UDP_Reader, error) {
	bts, err := x.Bind(port, ip)
	if err != nil {
		return nil, err
	}

	return &UDP_Reader{port: port, bytes: bts, manager: x, ipAddress: ip}, nil
}

func (c *UDP_Reader) Read(size int) ([]byte, error) {
	data := <-c.bytes
	if len(data) > size {
		data = data[:size]
	}
	return data, nil
}

func (c *UDP_Reader) Close() error {
	return c.manager.Unbind(c.port, c.ipAddress)
}
