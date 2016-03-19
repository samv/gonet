package udp

import (
	"github.com/hsheth2/gonet/ipv4"
)

type writer struct {
	rip      *ipv4.Address // destination ip address
	lip      *ipv4.Address // source ip address
	writer   ipv4.Writer
	src, dst Port // ports
}

func NewWriter(src, dest Port, dstIP *ipv4.Address) (Writer, error) {
	write, err := ipv4.NewWriter(dstIP, ipv4.IPProtoUDP)
	if err != nil {
		return nil, err
	}

	return &writer{
		src:    src,
		dst:    dest,
		rip:    dstIP,
		lip:    ipv4.GlobalRoutingTable.Query(dstIP),
		writer: write,
	}, nil
}

func (c *writer) Write(x []byte) (int, error) {
	headerLen := uint16(udpHeaderSize + len(x))
	UDPHeader := []byte{
		byte(c.src >> 8), byte(c.src), // Source port in byte slice
		byte(c.dst >> 8), byte(c.dst), // Destination port in byte slice
		byte(headerLen >> 8), byte(headerLen), // Length in bytes of UDP header + data
		0, 0, // Checksum
	}

	data := append(UDPHeader, x...)
	cksum := ipv4.CalcTransportChecksum(data, c.lip, c.rip, headerLen, ipv4.IPProtoUDP)
	data[6] = uint8(cksum >> 8)
	data[7] = uint8(cksum)

	///*logs*/logs.Trace.Println("UDP Writing:", data)
	return c.writer.WriteTo(data)
}

func (c *writer) Close() error {
	return nil
}
