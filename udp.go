package main

import (
    "fmt"
)

type UDP_manager struct {
    ipAddress string
    write *IP_Writer
    read  *IP_Reader
    buff      map[uint16](chan byte)
}

type UDP struct {
    manager   *UDP_manager
    bytes     chan byte
    src, dest uint16 // ports
}

func NewUDP_Manager(ip string) (*UDP_manager, error) {
    //p, err := net.ListenPacket("ip4:1", ip)
    //if err != nil {
    //    fmt.Println(err)
    //    return nil, err
    //}

    // TODO: Separate the server UDP and client UDP connections

    ipw, err := NewIP_Writer(ip)
    if err != nil {
        return nil, err
    }

    nr, err := NewNetwork_Reader()
    if err != nil {
        return nil, err
    }

    ipr, err := nr.NewIP_Reader(ip, 17) // 17 for UDP
    if err != nil {
        return nil, err;
    }

    x := &UDP_manager{
        read: ipr,
        write: ipw,
        buff: make(map[uint16](chan byte)),
        ipAddress: ip,
    }

    go x.readAll()

    return x, nil
}

func (x *UDP_manager) readAll() {
    for {
        _, payload, err := x.read.ReadFrom()
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
    return &UDP{src: src, dest: dest, bytes: x.buff[src], manager: x}, nil
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

    err := c.manager.write.WriteTo(x)
    if err != nil {
        fmt.Println(err)
        return err
    }
    return nil
}
func (c *UDP) close() error {
    delete(c.manager.buff, c.src)
    return nil
}
