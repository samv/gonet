package main

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	//"golang.org/x/net/ipv4"
)

type IP_Conn struct {
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

func NewIP_Conn(dst string) (*IP_Conn, error) {
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

	return &IP_Conn{
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

func calcChecksum(head []byte, excludeChecksum bool) uint16 {
	totalSum := uint64(0)
	for ind, elem := range head {
		if (ind == 10 || ind == 11) && excludeChecksum { // Ignore the checksum in some situations
			continue
		}

		if ind%2 == 0 {
			totalSum += (uint64(elem) << 8)
		} else {
			totalSum += uint64(elem)
		}
	}
	fmt.Println("Checksum total: ", totalSum)

	for prefix := (totalSum >> 16); prefix != 0; prefix = (totalSum >> 16) {
		//        fmt.Println(prefix)
		//        fmt.Println(totalSum)
		//        fmt.Println(totalSum & 0xffff)
		totalSum = uint64(totalSum&0xffff) + prefix
	}
	fmt.Println("Checksum after carry: ", totalSum)

	carried := uint16(totalSum)

	return ^carried
}

func slicePacket(b []byte) (hrd, payload []byte) {
	hdrLen := int(b[0]&0x0f) * 4
	fmt.Println("HdrLen: ", hdrLen)
	return b[:hdrLen], b[hdrLen:]
}

func (ipc *IP_Conn) ReadFrom(b []byte) (payload []byte, e error) {
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

func (ipc *IP_Conn) WriteTo(p []byte) error {
	totalLen := uint16(ipc.headerLen) + uint16(len(p))
	fmt.Println("Total Len: ", totalLen)
	packet := make([]byte, ipc.headerLen)
	packet[0] = (byte)((ipc.version << 4) + (uint8)(ipc.headerLen/4)) // Version, IHL
	packet[1] = 0
	packet[2] = (byte)(totalLen >> 8) // Total Len
	packet[3] = (byte)(totalLen)

	id := ipc.identifier
	packet[4] = byte(id >> 8) // Identification
	packet[5] = byte(id)
	ipc.identifier++

	packet[6] = byte(1 << 6)         // Flags: Don't fragment
	packet[7] = 0                    // Fragment Offset
	packet[8] = (byte)(ipc.ttl)      // Time to Live
	packet[9] = (byte)(ipc.protocol) // Protocol

	// Src and Dst IPs
	srcIP := net.ParseIP(ipc.src)
	fmt.Println(srcIP)
	//    fmt.Println(srcIP[12])
	//    fmt.Println(srcIP[13])
	//    fmt.Println(srcIP[14])
	//    fmt.Println(srcIP[15])
	dstIP := net.ParseIP(ipc.dst)
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

	//ipc.pc.WriteMsgIP(packet, nil, dstIPAddr)

	return syscall.Sendto(ipc.fd, packet, 0, ipc.sockAddr)
}

func (ipc *IP_Conn) Close() error {
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
