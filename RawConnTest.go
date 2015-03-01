package main

import (
    "fmt"
//    "net/ipv4"
    "net"
    "golang.org/x/net/ipv4"
    "os"
)

func main() {
    fmt.Println("Hello, World!")

    c, err := NewUDP(5049, 5050)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    c.read(1024)
    c.write(make([]byte, 0))
    c.close()
}

type UDP struct {
    open bool
	conn *ipv4.RawConn

    src, dest uint16

    pl net.PacketConn
}
func NewUDP(src, dest uint16) (*UDP, error) {
    p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", dest), "127.0.0.1")
    if err != nil {
        return nil, err;
    }

    r, err := ipv4.NewRawConn(p)
    if  err != nil {
        return nil, err;
    }

    // TODO use r.JoinGroup at https://godoc.org/golang.org/x/net/ipv4#NewRawConn

    return &UDP{open: true, conn: r, src: src, dest: dest, pl: p}, nil
}
func (c *UDP) read(size int) ([]byte, error) {
    return make([]byte, 0), nil
}
func (c *UDP) write(x []byte) error {
    // TODO convert to byte slice: 5 00 00 44 ad 0b 00 00 40 11 72 72 ac 14 02 fd ac 14 00 06
//	h := ipv4.ParseHeader()
//	c.conn.WriteTo(h, x, nil)
	return nil
}
func (c *UDP) close() error {
    return c.conn.Close()
}