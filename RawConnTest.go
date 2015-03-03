package main

import (
	"fmt"
	//    "net/ipv4"
	"golang.org/x/net/ipv4"
	"net"
	"os"
)

func main() {
	fmt.Println("Hello, World!")

	c, err := NewUDP(20001, 20005)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.read(1024)
	c.write([]byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', 0})
	c.close()
}

type UDP struct {
	open bool
	conn *ipv4.RawConn

	src, dest uint16

	pl net.PacketConn
}

func NewUDP(src, dest uint16) (*UDP, error) {
	p, err := net.ListenPacket("ip4:17", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r, err := ipv4.NewRawConn(p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// TODO use r.JoinGroup at https://godoc.org/golang.org/x/net/ipv4#NewRawConn

	return &UDP{open: true, conn: r, src: src, dest: dest, pl: p}, nil
}
func (c *UDP) read(size int) ([]byte, error) {
	return make([]byte, 0), nil
}
func (c *UDP) write(x []byte) error {
	UDPHeader := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte slice
		(byte)(c.dest >> 8), (byte)(c.dest), // Dest port in byte slice
		(byte)((8 + len(x)) >> 8), (byte)(8 + len(x)), // Length in bytes of UDP header + data
		0, 0, // Checksum
	}

	x = append(UDPHeader, x...)

	h := &ipv4.Header{
		Version:  ipv4.Version,      // protocol version
		Len:      20,                // header length
		TOS:      0,                 // type-of-service (0 is everything normal)
		TotalLen: len(x) + 20,       // packet total length (octets)
		ID:       0,                 // identification
		Flags:    ipv4.DontFragment, // flags
		FragOff:  0,                 // fragment offset
		TTL:      8,                 // time-to-live (maximum lifespan in seconds)
		Protocol: 17,                // next protocol (17 is UDP)
		Checksum: 0,                 // checksum (apparently autocomputed)
		//Src:    net.IPv4(127, 0, 0, 1), // source address, apparently done automatically
		Dst: net.IPv4(127, 0, 0, 1), // destination address
		//Options                         // options, extension headers
	}
	err := c.conn.WriteTo(h, x, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func (c *UDP) close() error {
	return c.conn.Close()
}
