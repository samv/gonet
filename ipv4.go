package main
import "net"

type IP_Conn struct {
    pc *net.PacketConn
    version uint8
    dst, src string
    headerLen uint16
    //len uint16
    //id uint16
    ttl uint8
    protocol uint8
    //checksum int
}

func NewIP_Conn(dst string) (*IP_Conn, error) {
    pc, _ := net.ListenPacket("ip4", dst)
    return &IP_Conn{
        pc: &pc,
        version: 4,
        headerLen: 20,
        dst:     dst,
        src:     "127.0.0.1",
        ttl:     8,
        protocol: 17,
    }, nil
}

func calcChecksum(head []byte, excludeChecksum bool) uint16 {
    totlSum := uint64(0)
    for ind, elem := range head {
        if (ind == 10 || ind == 11) && excludeChecksum { // Ignore the checksum in some situations
            continue
        }

        if ind % 2 == 0 {
            totlSum += (uint64)(uint16(elem) << 8)
        } else {
            totlSum += (uint64)(uint16(elem))
        }
    }

    for prefix := (totlSum >> 16); prefix != 0; {
        totlSum = uint64(uint16(totlSum)) + prefix
    }
    carried := uint16(totlSum)

    return ^carried
}

func (ipc *IP_Conn) ReadFrom() {

}

func (ipc *IP_Conn) WriteTo(p []byte) {
    totalLen := uint16(ipc.headerLen) + uint16(len(p))
    packet := make([]byte, totalLen)
    packet[0] = (byte)(ipc.version << 4) // Version
    packet[1] = 0
    packet[2] = (byte)(totalLen >> 8) // Total Len
    packet[3] = (byte)(totalLen)
    packet[4] = 0 // Identification (for now)
    packet[5] = 0
    packet[6] = byte(1 << 6) // Flags: Don't fragment
    packet[7] = 0 // Fragment Offset
    packet[8] = (byte)(ipc.ttl) // Time to Live
    packet[9] = (byte)(ipc.protocol) // Protocol

    // Src and Dst IPs
    srcIP := net.ParseIP(ipc.src)
    dstIP := net.ParseIP(ipc.dst)
    packet[12] = srcIP[0]
    packet[13] = srcIP[1]
    packet[14] = srcIP[2]
    packet[15] = srcIP[3]
    packet[16] = dstIP[0]
    packet[17] = dstIP[1]
    packet[18] = dstIP[2]
    packet[19] = dstIP[3]

    // Checksum
    checksum := calcChecksum(packet[:20], true)
    packet[10] = byte(checksum >> 8)
    packet[11] = byte(checksum)

    // Payload
    for ind, elem := range p {
        packet[20+ind] = elem
    }
    
//    ipc.pc.WriteTo
}

func (ipc *IP_Conn) Close() error {
    return nil
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
