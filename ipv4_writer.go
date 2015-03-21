package main

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

type IP_Writer struct {
	fd         int
	sockAddr   syscall.Sockaddr
	version    uint8
	dst, src   string
	headerLen  uint16
	ttl        uint8
	protocol   uint8
	identifier uint16
}

func NewIP_Writer(dst string, protocol uint8) (*IP_Writer, error) {
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

	return &IP_Writer{
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

func (ipw *IP_Writer) WriteTo(p []byte) error {
	header := make([]byte, ipw.headerLen)
	header[0] = (byte)((ipw.version << 4) + (uint8)(ipw.headerLen/4)) // Version, IHL
	header[1] = 0
	id := ipw.identifier
	header[4] = byte(id >> 8) // Identification
	header[5] = byte(id)
	ipw.identifier++
	header[6] = byte(1 << 5)         // Flags: May fragment, more fragments
	header[8] = (byte)(ipw.ttl)      // Time to Live
	header[9] = (byte)(ipw.protocol) // Protocol

	// Src and Dst IPs
	srcIP := net.ParseIP(ipw.src)
	fmt.Println(srcIP)
	//    fmt.Println(srcIP[12])
	//    fmt.Println(srcIP[13])
	//    fmt.Println(srcIP[14])
	//    fmt.Println(srcIP[15])
	dstIP := net.ParseIP(ipw.dst)
	fmt.Println(dstIP)
	header[12] = srcIP[12]
	header[13] = srcIP[13]
	header[14] = srcIP[14]
	header[15] = srcIP[15]
	header[16] = dstIP[12]
	header[17] = dstIP[13]
	header[18] = dstIP[14]
	header[19] = dstIP[15]

	for i := 0; i < len(p)/1480+1; i += 1480 {
		if len(p) <= 1480*(i+1) {
			header[6] = byte(0)
		}
		// TODO allow frag offset to be full 13 bits instead of current 8 (needs to use packet[6] as well
		p[7] = byte(i) // Fragment offset

		totalLen := uint16(ipw.headerLen) + uint16(len(p))
		fmt.Println("Total Len: ", totalLen)
		header[2] = (byte)(totalLen >> 8) // Total Len
		header[3] = (byte)(totalLen)

		// IPv4 header test (before checksum)
		fmt.Println("Packet before checksum: ", header)
		// Checksum
		checksum := calcChecksum(header[:20], true)
		header[10] = byte(checksum >> 8)
		header[11] = byte(checksum)
		// Payload

		newPacket := make([]byte, 1)
		if len(p) <= 1480*(i+1) {
			header[6] = byte(0)
			newPacket = append(header, p[1480*i:]...)
			fmt.Println("Full Packet:  ", newPacket)
		} else {
			newPacket = append(header, p[1480*i:1480*(i+1)]...)
			fmt.Println("Full Packet:  ", newPacket)
		}

		err := syscall.Sendto(ipw.fd, newPacket, 0, ipw.sockAddr)
		if err != nil {
			return err
		}
	}

	// TODO: Allow IP fragmentation (use 1500 as MTU)
	return nil
}

func (ipw *IP_Writer) Close() error {
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
