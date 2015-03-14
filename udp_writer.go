package main

//import (
//    "fmt"
//)

type UDP_Writer struct {
	ipAddress string // destination ip address
	writer    *IP_Writer
	src, dst  uint16 // ports
}

func NewUDP_Writer(src, dest uint16, dstIP string) (*UDP_Writer, error) {
	write, err := NewIP_Writer(dstIP, 17) // 17 for UDP
	if err != nil {
		return nil, err
	}

	return &UDP_Writer{src: src, dst: dest, ipAddress: dstIP, writer: write}, nil
}

func (c *UDP_Writer) write(x []byte) error {
	UDPHeader := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte slice
		(byte)(c.dst >> 8), (byte)(c.dst), // Destination port in byte slice
		(byte)((8 + len(x)) >> 8), (byte)(8 + len(x)), // Length in bytes of UDP header + data
		0, 0, // Checksum
	}

	x = append(UDPHeader, x...)

	err := c.writer.WriteTo(x)
	if err != nil {
		//fmt.Println(err)
		return err
	}
	return nil
}
func (c *UDP_Writer) close() error {
	return nil
}
