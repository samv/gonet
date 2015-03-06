package main
import "net"

type IP_Conn struct {
    pc *net.PacketConn
    version uint8
    dst, src string
    //len uint16
    //id uint16
    ttl int
    protocol uint8
    //checksum int
}

func NewIP_Conn(dst string) (*IP_Conn, error) {
    pc, _ := net.ListenPacket("ip4", dst)
    return &IP_Conn{
        pc: &pc,
        version: 4,
        dst:     dst,
        src:     "127.0.0.1",
        ttl:     8,
        protocol: 17,
    }, nil
}

func (ipc *IP_Conn) ReadFrom() {

}

func (ipc *IP_Conn) WriteTo(p []byte) {

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
