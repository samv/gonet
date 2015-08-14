package udp

import (
	"network/ipv4"
	//"github.com/hsheth2/logs"
)

const UDP_HEADER_SZ = 8

type UDP_Writer struct {
	rip      ipv4.IPaddress // destination ip address
	lip      ipv4.IPaddress // source ip address
	writer   *ipv4.IP_Writer
	src, dst uint16 // ports
}

func NewUDP_Writer(src, dest uint16, dstIP ipv4.IPaddress) (*UDP_Writer, error) {
	write, err := ipv4.NewIP_Writer(dstIP, ipv4.UDP_PROTO)
	if err != nil {
		return nil, err
	}

	return &UDP_Writer{
		src:    src,
		dst:    dest,
		rip:    dstIP,
		lip:    ipv4.GlobalSource_IP_Table.Query(dstIP),
		writer: write,
	}, nil
}

func (c *UDP_Writer) Write(x []byte) error {
	headerLen := uint16(UDP_HEADER_SZ + len(x))
	UDPHeader := []byte{
		byte(c.src >> 8), byte(c.src), // Source port in byte slice
		byte(c.dst >> 8), byte(c.dst), // Destination port in byte slice
		byte(headerLen >> 8), byte(headerLen), // Length in bytes of UDP header + data
		0, 0, // Checksum
	}

	data := append(UDPHeader, x...)
	cksum := ipv4.CalcTransportChecksum(data, c.lip, c.rip, headerLen, ipv4.UDP_PROTO)
	data[6] = uint8(cksum >> 8)
	data[7] = uint8(cksum)

	//logs.Trace.Println("UDP Writing:", data)
	return c.writer.WriteTo(data)
}

func (c *UDP_Writer) Close() error {
	return nil
}
