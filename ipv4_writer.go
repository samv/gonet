package main

import (
	"fmt"
	"net"
	"syscall"
	//"errors"
)

type IP_Writer struct {
	fd          int
	sockAddr    syscall.Sockaddr
	version     uint8
	dst, src    string
	headerLen   uint16
	ttl         uint8
	protocol    uint8
	identifier  uint16
	maxFragSize uint16
}

func NewIP_Writer(dst string, protocol uint8) (*IP_Writer, error) {
	fd, err := syscall.Socket(AF_PACKET, SOCK_RAW, HTONS_ETH_P_ALL)
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

	addr := &syscall.SockaddrLinklayer{
        //Protocol: 0x0800, //IP
        // Family is AF_PACKET
        Protocol: HTONS_ETH_P_IP, // should be inherited anyway
		Addr: myMACAddr, // sending to myself
        Halen: ETH_ALEN, // may not be correct
        Ifindex: MyIfIndex, // TODO: don't hard code this... fix it later
	}

	err = syscall.Sendto(fd, []byte{0x08, 0x00, 0x27, 0x9e, 0x29, 0x63, 0x08, 0x00, 0x27, 0x9e, 0x29, 0x63, 0x08, 0x00}, 0, addr) //Random bytes
	if err != nil {
		fmt.Println("ERROR returned by syscall.Sendto", err)
	} else {
        fmt.Println("Send the test packet")
    }

	/*err = syscall.Connect(fd, addr)
	if err != nil {
		return nil, errors.New("Failed to connect.")
	}*/

	return &IP_Writer{
		fd:          fd,
		sockAddr:    addr,
		version:     4,
		headerLen:   20,
		dst:         dst,
		src:         "127.0.0.1",
		ttl:         8,
		protocol:    17,
		identifier:  20000,
		maxFragSize: 1500,
	}, nil
}

func (ipw *IP_Writer) WriteTo(p []byte) error {
    // build the ethernet header
    /*etherHead :=  append(append(
        myMACSlice, // dst MAC
        myMACSlice...), // src MAC
        0x08, 0x00, // ethertype (IP)
    )*/
    etherHead := []byte{
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0,
    }
    fmt.Println("My header:", etherHead)

	header := make([]byte, ipw.headerLen)
	header[0] = (byte)((ipw.version << 4) + (uint8)(ipw.headerLen/4)) // Version, IHL
	header[1] = 0
	id := ipw.identifier
	header[4] = byte(id >> 8) // Identification
	header[5] = byte(id)
	ipw.identifier++
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

	maxFragSize := int(ipw.maxFragSize)
	maxPaySize := maxFragSize - int(ipw.headerLen)

	for i := 0; i < len(p)/maxPaySize+1; i += 1 {
		fmt.Println("Looping fragmenting")
		if len(p) <= maxPaySize*(i+1) {
			header[6] = byte(0)
		} else {
			header[6] = byte(1 << 5) // Flags: May fragment, more fragments
		}
		fmt.Println("off", i*maxFragSize, byte((i*maxFragSize)>>8), byte(i*maxFragSize))

		offset := (i * maxPaySize) / 8
		fmt.Println("Header 6 before:", header[6])
		header[6] += byte(offset >> 8)
		fmt.Println("Header 6 after:", header[6])
		header[7] = byte(offset) // Fragment offset

		totalLen := uint16(0)

		// Payload
		newPacket := make([]byte, 1)
		if len(p) <= maxFragSize*(i+1) {
			totalLen = uint16(ipw.headerLen) + uint16(len(p[maxPaySize*i:]))
			fmt.Println("Full Pack")
			fmt.Println("len", len(p[maxPaySize*i:]))
			//header[6] = byte(0)

			fmt.Println("Total Len: ", totalLen)
			header[2] = (byte)(totalLen >> 8) // Total Len
			header[3] = (byte)(totalLen)

			// IPv4 header test (before checksum)
			fmt.Println("Packet before checksum: ", header)
			// Checksum
			checksum := calculateChecksum(header[:20])
			header[10] = byte(checksum >> 8)
			header[11] = byte(checksum)

			newPacket = append(header, p[maxPaySize*i:]...)
			fmt.Println("Full Packet:  ", newPacket)
			fmt.Println("CALCULATED LEN:", i*maxFragSize+len(p[maxPaySize*i:]))
		} else {
			totalLen = uint16(ipw.headerLen) + uint16(len(p[maxPaySize*i:maxPaySize*(i+1)]))
			fmt.Println("Partial packet")
			fmt.Println("len", len(p[maxPaySize*i:maxPaySize*(i+1)]))

			fmt.Println("Total Len: ", totalLen)
			header[2] = (byte)(totalLen >> 8) // Total Len
			header[3] = (byte)(totalLen)

			// IPv4 header test (before checksum)
			fmt.Println("Packet before checksum: ", header)
			// Checksum
			checksum := calculateChecksum(header[:20])
			header[10] = byte(checksum >> 8)
			header[11] = byte(checksum)

			newPacket = append(header, p[maxPaySize*i:maxPaySize*(i+1)]...)
			fmt.Println("Full Packet:  ", newPacket)
		}

        // add on the ethernet header
        newPacket = append(etherHead, newPacket...)
        fmt.Println("Full Packet with ethernet header:", newPacket)

		err := syscall.Sendto(ipw.fd, newPacket, 0, ipw.sockAddr)
		if err != nil {
			return err
		}
	}
	fmt.Println("PAY LEN", len(p))

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
