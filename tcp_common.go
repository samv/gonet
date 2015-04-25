package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
)

// Finite State Machine
const (
	CLOSED      = 1
	LISTEN      = 2
	SYN_SENT    = 3
	SYN_RCVD    = 4
	ESTABLISHED = 5
	FIN_WAIT_1  = 6
	FIN_WAIT_2  = 7
	CLOSE_WAIT  = 8
	CLOSING     = 9
	LAST_ACK    = 10
	TIME_WAIT   = 11

	FSM_NUM_STATES = 11
)

// TCB Types
const (
	TCP_SERVER = 1
	TCP_CLIENT
)

// Other Consts
const TCP_INCOMING_BUFF_SZ = 10
const TCP_BASIC_HEADER_SZ = 20
const TCP_LISTEN_QUEUE_SZ = 10
const TCP_RESEND_LIMIT = 12

// Window Sizing
const MAX_WINDOW_SZ = 65000
const MIN_WINDOW_SZ = 500

// TODO: set these properly based on the standard values

// Flags
const (
	TCP_FIN = 0x01
	TCP_SYN = 0x02
	TCP_RST = 0x04
	TCP_PSH = 0x08
	TCP_ACK = 0x10
	TCP_URG = 0x20
	TCP_ECE = 0x40
	TCP_CWR = 0x80
)

// Global src, dst port and ip registry for TCP binding
type TCP_Port_Manager_Type struct {
	tcp_reader *IP_Reader
	incoming   map[uint16](map[uint16](map[string](chan *TCP_Packet))) // dst, src port, remote ip
}

// TODO TCP_Port_Manager_Type should have an unbind function
func (m *TCP_Port_Manager_Type) bind(srcport, dstport uint16, ip string) (chan *TCP_Packet, error) {
	// dstport is the local one here, srcport is the remote
	if _, ok := m.incoming[dstport]; !ok {
		m.incoming[dstport] = make(map[uint16](map[string](chan *TCP_Packet)))
	}

	// TODO add an option (for servers) for all srcports
	if _, ok := m.incoming[dstport][srcport]; !ok {
		m.incoming[dstport][srcport] = make(map[string](chan *TCP_Packet))
	}

	if _, ok := m.incoming[dstport][srcport][ip]; ok {
		return nil, errors.New("Ports and IP already binded to")
	}

	m.incoming[dstport][srcport][ip] = make(chan *TCP_Packet, TCP_INCOMING_BUFF_SZ)
	return m.incoming[dstport][srcport][ip], nil
}

func (m *TCP_Port_Manager_Type) readAll() {
	for {
		rip, lip, _, payload, err := m.tcp_reader.ReadFrom()
		if err != nil {
			fmt.Println("TCP readAll error", err) // TODO log instead of print
			continue
		}

		p, err := Extract_TCP_Packet(payload, rip, lip)
		if err != nil {
			fmt.Println(err)
			continue
		}

		rport := p.header.srcport
		lport := p.header.dstport

		var output chan *TCP_Packet = nil

		if _, ok := m.incoming[lport]; ok {
			if _, ok := m.incoming[lport][rport]; ok {
				if _, ok := m.incoming[lport][rport][rip]; ok {
					output = m.incoming[lport][rport][rip]
				} else if _, ok := m.incoming[lport][rport]["*"]; ok {
					output = m.incoming[lport][rport]["*"]
				}
			}
			if _, ok := m.incoming[lport][0]; ok {
				if _, ok := m.incoming[lport][0]["*"]; ok {
					output = m.incoming[lport][0]["*"]
				}
			}
		}

		if output != nil {
			go func() { output <- p }()
		} else {
			// TODO send a rst to sender if nothing is binded to the dst port, src port, and remote ip
			//fmt.Println(errors.New("Dst/Src port + ip not binded to"))
		}
	}
}

var TCP_Port_Manager = func() *TCP_Port_Manager_Type {
	nr, err := NewNetwork_Reader() // TODO: create a global var for the network reader
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ipr, err := nr.NewIP_Reader("*", TCP_PROTO)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	m := &TCP_Port_Manager_Type{
		tcp_reader: ipr,
		incoming:   make(map[uint16](map[uint16](map[string](chan *TCP_Packet)))),
	}
	go m.readAll()
	return m
}()

type TCP_Packet struct {
	header   *TCP_Header
	payload  []byte
	rip, lip string
}

func (p *TCP_Packet) Marshal_TCP_Packet() ([]byte, error) {
	head, err := p.header.Marshal_TCP_Header(p.rip, p.lip)
	packet := append(head, p.payload...)
	return packet, err
}

func (p *TCP_Packet) getPayloadSize() uint32 {
	if len(p.payload) == 0 {
		return 1
	}
	return uint32(len(p.payload))
}

type TCP_Header struct {
	srcport, dstport uint16
	seq, ack         uint32
	// will do data offset automatically
	flags  uint8
	window uint16
	// checksum will be automatic
	urg     uint16
	options []byte
}

func (h *TCP_Header) Marshal_TCP_Header(dstIP, srcIP string) ([]byte, error) {
	// pad options with 0's
	for len(h.options)%4 != 0 {
		h.options = append(h.options, 0)
	}

	headerLen := uint8(TCP_BASIC_HEADER_SZ + len(h.options)) // size of header in 32 bit (4 byte) chunks

	header := append([]byte{
		(byte)(h.srcport >> 8), (byte)(h.srcport), // Source port in byte slice
		(byte)(h.dstport >> 8), (byte)(h.dstport), // Destination port in byte slice
		(byte)(h.seq >> 24), (byte)(h.seq >> 16), (byte)(h.seq >> 8), (byte)(h.seq), // seq
		(byte)(h.ack >> 24), (byte)(h.ack >> 16), (byte)(h.ack >> 8), (byte)(h.ack), // ack
		(byte)(
			(headerLen / 4) << 4, // data offset.
		// bits 5-7 inclusive are reserved, always 0
		// bit 8 is flag 0(NS flag), set to 0 here because only SYN
		),
		byte(h.flags),
		byte(h.window >> 8), byte(h.window), // window
		0, 0, // checksum (0 for now, set later)
		byte(h.urg >> 8), byte(h.urg), // URG pointer, only matters where URG flag is set
	}, h.options...)

	// insert the checksum
	cksum := calcTCPchecksum(header, srcIP, dstIP, headerLen)
	header[16] = byte(cksum >> 8)
	header[17] = byte(cksum)

	return header, nil
}

func Extract_TCP_Packet(d []byte, rip, lip string) (*TCP_Packet, error) {
	// TODO: test this function fully

	// header length
	headerLen := (d[12] >> 4) * 4
	if headerLen < TCP_BASIC_HEADER_SZ {
		return nil, errors.New("Bad TCP header size: Less than 20.")
	}

	// checksum verification
	if !verifyTCPchecksum(d[:headerLen], rip, lip, headerLen) {
		return nil, errors.New("Bad TCP header checksum")
	}

	// create the header
	h := &TCP_Header{
		srcport: uint16(d[0])<<8 | uint16(d[1]),
		dstport: uint16(d[2])<<8 | uint16(d[3]),
		seq:     uint32(d[4])<<24 | uint32(d[5])<<16 | uint32(d[6])<<8 | uint32(d[7]),
		ack:     uint32(d[8])<<24 | uint32(d[9])<<16 | uint32(d[10])<<8 | uint32(d[11]),
		flags:   uint8(d[13]),
		window:  uint16(d[14])<<8 | uint16(d[15]),
		urg:     uint16(d[18])<<8 | uint16(d[19]),
		options: d[TCP_BASIC_HEADER_SZ:headerLen],
	}
	return &TCP_Packet{header: h, payload: d[headerLen:], rip: rip, lip: lip}, nil
}

func calcTCPchecksum(header []byte, srcIP, dstIP string, headerLen uint8) uint16 {
	return checksum(append(append(append(header, net.ParseIP(srcIP)...), net.ParseIP(dstIP)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(headerLen >> 8), byte(headerLen)}...))
}

func verifyTCPchecksum(header []byte, srcIP, dstIP string, headerLen uint8) bool {
	// TODO: do TCP checksum verification
	return true
}

func genRandSeqNum() uint32 {
	x := make([]byte, 4) // four bytes
	_, err := rand.Read(x)
	if err != nil {
		fmt.Println(errors.New("Failed to genRandSeqNum")) // TODO log instead of print
		return 0                                           // TODO incorporate an error message
	}
	return uint32(x[0])<<24 | uint32(x[1])<<16 | uint32(x[2])<<8 | uint32(x[3])
}

func MyRawConnTCPWrite(w *ipv4.RawConn, tcp []byte, dst string) error {
	return w.WriteTo(&ipv4.Header{
		Version:  ipv4.Version,             // protocol version
		Len:      IP_HEADER_LEN,            // header length
		TOS:      0,                        // type-of-service (0 is everything normal)
		TotalLen: len(tcp) + IP_HEADER_LEN, // packet total length (octets)
		ID:       0,                        // identification
		Flags:    ipv4.DontFragment,        // flags
		FragOff:  0,                        // fragment offset
		TTL:      DEFAULT_TTL,              // time-to-live (maximum lifespan in seconds)
		Protocol: TCP_PROTO,                // next protocol
		Checksum: 0,                        // checksum (autocomputed)
		Dst:      net.ParseIP(dst),         // destination address
	}, tcp, nil)
}