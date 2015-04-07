package main

type TCP_Client struct {
	ipAddress string // destination ip address
	writer    *IP_Writer
	src, dst  uint16 // ports
}

func New_TCP_Client(src, dest uint16, dstIP string) (*TCP_Client, error) {
	write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}

	return &TCP_Client{src: src, dst: dest, ipAddress: dstIP, writer: write}, nil
}

func (c *TCP_Client) Connect() error {
	// TODO Send SYN
	// TODO Wait for SYN + ACK, send back ACK
	return nil
}

func (c *TCP_Client) Send(data []byte) error {
	return nil
}

func (c *TCP_Client) Close() error {
	return nil
}
