package oldIPv4

import (
    "fmt"
    "net"
)

type UDP_manager struct {
    ipAddress string
//    pl        net.PacketConn
    open      bool
    conn      *IP_Conn
    buff      map[uint16](chan byte)
}

type UDP struct {
    manager   *UDP_manager
    conn      *IP_Conn
    bytes     chan byte
    src, dest uint16
}

func NewUDP_Manager(ip string) (*UDP_manager, error) {
//    p, err := net.ListenPacket("ip4:1", ip)
//    if err != nil {
//        fmt.Println(err)
//        return nil, err
//    }

    r, err := NewIP_Conn(ip)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }

    x := &UDP_manager{open: true, conn: r, /*pl: p,*/ buff: make(map[uint16](chan byte)), ipAddress: ip}

    go x.readAll()

    return x, nil
}

func (x *UDP_manager) readAll() {
    b := make([]byte, 1024)

    for {
        payload, err := x.conn.ReadFrom(b)
        if err != nil {
            continue
        }
        //        fmt.Println(b)
        fmt.Println("UDP header and payload: ", payload)

        dest := (((uint16)(payload[2])) * 256) + ((uint16)(payload[3]))
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

    err := c.conn.WriteTo(x)
    if err != nil {
        fmt.Println(err)
        return err
    }
    return nil
}
func (c *UDP) close() error {
    return c.conn.Close()
}
