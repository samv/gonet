package main

import (
    "golang.org/x/net/ipv4"
    "net"
    "fmt"
)


type UDP_manager struct {
    ipAddress string
    pl        net.PacketConn
    open      bool
    conn      *ipv4.RawConn
    buff      map[uint16](chan byte)
}

type UDP struct {
    manager   *UDP_manager
    conn      *ipv4.RawConn
    bytes     chan byte
    src, dest uint16
}

func NewUDP_Manager(ip string) (*UDP_manager, error) {
    p, err := net.ListenPacket("ip4:17", ip)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    r, err := ipv4.NewRawConn(p)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    x := &UDP_manager{open: true, conn: r, pl: p, buff: make(map[uint16](chan byte)), ipAddress: ip}

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
        //		fmt.Println(dest)
        //		fmt.Println(payload)
        //
        //		fmt.Println(x.buff)
        c, ok := x.buff[dest]
        //fmt.Println(ok)
        payload = payload[8:]
        if ok {
            go func() {
                for _, elem := range payload {
                    //fmt.Println("Writing")
                    c <- elem
                }
            }()
        }
    }
}

func (x *UDP_manager) NewUDP(src, dest uint16) (*UDP, error) {
    x.buff[src] = make(chan byte, 1024)
    return &UDP{src: src, dest: dest, conn: x.conn, bytes: x.buff[src], manager: x}, nil
}

func (c *UDP) read(size int) ([]byte, error) {
    data := make([]byte, size)
    for i := 0; i < size; i++ {
        //fmt.Println("test")
        data[i] = <-c.bytes
        //fmt.Println(data[i])
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
        Version:  ipv4.Version, // protocol version
        Len:      20, // header length
        TOS:      0, // type-of-service (0 is everything normal)
        TotalLen: len(x) + 20, // packet total length (octets)
        ID:       0, // identification
        Flags:    ipv4.DontFragment, // flags
        FragOff:  0, // fragment offset
        TTL:      8, // time-to-live (maximum lifespan in seconds)
        Protocol: 17, // next protocol (17 is UDP)
        Checksum: 0, // checksum (apparently autocomputed)
        //Src:    net.IPv4(127, 0, 0, 1), // source address, apparently done automatically
        Dst: net.ParseIP(c.manager.ipAddress), // destination address
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

