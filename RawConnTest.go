package main

import (
	"fmt"
	//    "net/ipv4"
	"golang.org/x/net/ipv4"
	"net"
	"os"
)

func main() {
    manager, _ := NewUDP_Manager();
	c, err := NewUDP(20006, 20005, manager)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(c.read(41))
	c.write([]byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', 0})
	c.close()
}

type UDP_manager struct {
    pl net.PacketConn
    open bool
    conn *ipv4.RawConn
}

type UDP struct {
    conn *ipv4.RawConn
	src, dest uint16
}

func NewUDP_Manager() (*UDP_manager, error) {
	p, err := net.ListenPacket("ip4:17", "127.0.0.1")
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

	return &UDP_manager{open: true, conn: r, pl: p}, nil
}
func NewUDP(src, dest uint16, manager *UDP_manager) (*UDP, error) {
    return &UDP{src: src, dest: dest, conn: manager.conn}, nil
}

func (c *UDP) read(size int) ([]byte, error) {
	b := make([]byte, size)
	_, payload, _, err := c.conn.ReadFrom(b)
	if err != nil {
		return nil, err
	}
	return payload[8:], nil
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
