package main

import (
    "errors"
    "fmt"
    "net"
    "syscall"
)

type IP_Client struct {
    fd        int
    sockAddr  syscall.Sockaddr
    version   uint8
    dst, src  string
    headerLen uint16
    ttl      uint8
    protocol uint8
    identifier uint16
}

func NewIP_Client(dst string) (*IP_Client, error) {
    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
    if err != nil {
        fmt.Println("Write's socket failed")
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
        Addr: [4]byte{
            dstIPAddr.IP[12],
            dstIPAddr.IP[13],
            dstIPAddr.IP[14],
            dstIPAddr.IP[15],
        },
    }

    err = syscall.Connect(fd, addr)
    if err != nil {
        return nil, errors.New("Failed to connect.")
    }

    return &IP_Client{
        fd:         fd,
        sockAddr:   addr,
        version:    4,
        headerLen:  20,
        dst:        dst,
        src:        "127.0.0.1",
        ttl:        8,
        protocol:   17,
        identifier: 20000,
    }, nil
}

func (ipw *IP_Client) WriteTo(p []byte) error {
    totalLen := uint16(ipw.headerLen) + uint16(len(p))
    fmt.Println("Total Len: ", totalLen)
    packet := make([]byte, ipw.headerLen)
    packet[0] = (byte)((ipw.version << 4) + (uint8)(ipw.headerLen/4)) // Version, IHL
    packet[1] = 0
    packet[2] = (byte)(totalLen >> 8) // Total Len
    packet[3] = (byte)(totalLen)

    id := ipw.identifier
    packet[4] = byte(id >> 8) // Identification
    packet[5] = byte(id)
    ipw.identifier++

    packet[6] = byte(1 << 6)         // Flags: Don't fragment
    packet[7] = 0                    // Fragment Offset
    packet[8] = (byte)(ipw.ttl)      // Time to Live
    packet[9] = (byte)(ipw.protocol) // Protocol

    // Src and Dst IPs
    srcIP := net.ParseIP(ipw.src)
    fmt.Println(srcIP)
    //    fmt.Println(srcIP[12])
    //    fmt.Println(srcIP[13])
    //    fmt.Println(srcIP[14])
    //    fmt.Println(srcIP[15])
    dstIP := net.ParseIP(ipw.dst)
    fmt.Println(dstIP)
    packet[12] = srcIP[12]
    packet[13] = srcIP[13]
    packet[14] = srcIP[14]
    packet[15] = srcIP[15]
    packet[16] = dstIP[12]
    packet[17] = dstIP[13]
    packet[18] = dstIP[14]
    packet[19] = dstIP[15]

    // IPv4 header test (before checksum)
    fmt.Println("Packet before checksum: ", packet)

    // Checksum
    checksum := calcChecksum(packet[:20], true)
    packet[10] = byte(checksum >> 8)
    packet[11] = byte(checksum)

    // Payload
    packet = append(packet, p...)
    fmt.Println("Full Packet:  ", packet)

    return syscall.Sendto(ipw.fd, packet, 0, ipw.sockAddr)
}

func (ipw *IP_Client) Close() error {
    return syscall.Close(ipw.fd)
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
