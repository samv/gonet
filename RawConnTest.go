package main

import (
	"fmt"
	//    "net/ipv4"
	"golang.org/x/net/ipv4"
	"net"
	"os"
)

func main() {
	manager, _ := NewUDP_Manager()
	c, err := manager.NewUDP(20006, 20005)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(c.read(41))
	c.write([]byte{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', 0})
	c.close()
}

type UDP_manager struct {
	pl   net.PacketConn
	open bool
	conn *ipv4.RawConn
    buff map[uint16](chan byte)
}

type UDP struct {
    conn      *ipv4.RawConn
    bytes chan byte
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

    x := &UDP_manager{open: true, conn: r, pl: p, buff: make(map[uint16](chan byte))}

    go x.readAll()

    return x, nil
}

func (x *UDP_manager) readAll() {
    b := make([]byte, 1024)

    for {
        _, payload, _, err := x.conn.ReadFrom(b)
        if err != nil {
            continue
        }

        dest := (((uint16)(payload[2])) << 8) + ((uint16)(payload[3]))

        c, ok := x.buff[dest]
        if ok {
            go func() {
                for _, elem := range(payload) {
                    c <- elem
                }
            }()
        }
    }
}

func (x *UDP_manager) NewUDP(src, dest uint16) (*UDP, error) {
    x.buff[dest] = make(chan byte, 1024)
    return &UDP{src: src, dest: dest, conn: x.conn}, nil
}

func (c *UDP) read(size int) ([]byte, error) {
    data := make([]byte, size)
	for i := 0; i < size; i++ {
        data[i] = <-c.bytes
    }
    return data, nil
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
