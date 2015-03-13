package main

import (
    "errors"
    "fmt"
    "net"
    "syscall"
//"golang.org/x/net/ipv4"
)

type IP_Reader struct {
    fd        int
    sockAddr  syscall.Sockaddr
    version   uint8
    dst, src  string
    headerLen uint16
    //len uint16
    //id uint16
    ttl      uint8
    protocol uint8
    //checksum int
    identifier uint16
}

func NewIP_Reader(dst string) (*IP_Reader, error) {
    //pc, err := net.ListenIP("ip4:17", &net.IPAddr{IP: net.ParseIP(dst)})
    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
    if err != nil {
        fmt.Println("Failed to ListenIP")
        return nil, err
    }

    dstIPAddr, err := net.ResolveIPAddr("ip", dst)
    if err != nil {
        //fmt.Println(err)
        return nil, err
    }
    fmt.Println("Full Address: ", dstIPAddr)

    addr := &syscall.SockaddrInet4{
        Port: 20000,
        //Addr: [4]byte{127, 0, 0, 1},
        Addr: [4]byte{
            dstIPAddr.IP[12],
            dstIPAddr.IP[13],
            dstIPAddr.IP[14],
            dstIPAddr.IP[15],
        },
    }

    /*err = syscall.Connect(fd, addr)
    if err != nil {
        return nil, errors.New("Failed to connect.")
    }*/
    err = syscall.Bind(fd, addr)
    if err != nil {
        return nil, errors.New("Failed to bind to address.")
    }

    return &IP_Reader{
        fd:         fd,
        sockAddr:   addr,
        version:    4,
        headerLen:  20,
        dst:        dst,
        src:        "127.0.0.1",
        ttl:        8,
        protocol:   17,
    }, nil
}

func slicePacket(b []byte) (hrd, payload []byte) {
    hdrLen := int(b[0]&0x0f) * 4
    fmt.Println("HdrLen: ", hdrLen)
    return b[:hdrLen], b[hdrLen:]
}

func (ipc *IP_Reader) ReadFrom(b []byte) (payload []byte, e error) {
    //n, _, err := syscall.Recvfrom(ipc.fd, b, 0) //_ is src address
    n, _, _, _, err := syscall.Recvmsg(ipc.fd, b, make([]byte, 30000), 0)
    b = b[:n]
    fmt.Println("Read Length: ", n)
    fmt.Println("Full Read Data (after trim): ", b)
    hdr, p := slicePacket(b)

    // verify checksum
    if calcChecksum(hdr, false) != 0 {
        fmt.Println("Header checksum verification failed. Packet dropped.")
        fmt.Println("Wrong header: ", hdr)
        fmt.Println("Payload (dropped): ", p)
        return nil, errors.New("Header checksum incorrect, packet dropped")
    }

    return p, err
}

func (ipc *IP_Reader) Close() error {
    return syscall.Close(ipc.fd)
}

/* h := &ipv4.Header{
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
	Dst: net.ParseIP(c.manager.ipAddress), // destination address
	//Options                         // options, extension headers
}
*/
