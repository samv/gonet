package main

import (
	"net"
	"golang.org/x/net/ipv4"
)

// Finite State Machine
const (
	LISTEN = 1
	SYN_SENT
	SYN_RCVD
	ESTABLISHED
	FIN_WAIT_1
	FIN_WAIT_2
	CLOSE_WAIT
	CLOSING
	LAST_ACK
	TIME_WAIT
	CLOSED
)

// TCB Types
const (
	TCP_SERVER = 1
	TCP_CLIENT
)

// Other Consts
const TCP_INCOMING_BUFF_SZ = 10
const TCP_BASIC_HEADER_SZ = 20

// Flags
const (
	TCP_FIN = 0x01
	TCP_SYN = 0x02
	TCP_RSH = 0x04
	TCP_PSH = 0x08
	TCP_ACK = 0x10
	TCP_URG = 0x20
	TCP_ECE = 0x40
	TCP_CWR = 0x80
)

type TCP_Header struct {
	srcport, dstport uint16
	seq, ack uint32
	// will do data offset automatically
	flags uint8
	window uint16
	// checksum will be automatic
	urg uint16
	options []byte
}

func Make_TCP_Header(h *TCP_Header, dstIP, srcIP string) ([]byte, error) {
	// pad options with 0's
	for len(h.options) % 4 != 0 {
		h.options = append(h.options, 0)
	}

	headerLen := uint8(TCP_BASIC_HEADER_SZ + len(h.options)) // size of header in 32 bit (4 byte) chunks

	header := append([]byte{
		(byte)(h.srcport >> 8),     (byte)(h.srcport), // Source port in byte slice
		(byte)(h.dstport >> 8),     (byte)(h.dstport), // Destination port in byte slice
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

	cksum := calcTCPchecksum(header, srcIP, dstIP, headerLen)
	header[16] = byte(cksum >> 8)
	header[17] = byte(cksum)

	return header, nil
}

func calcTCPchecksum(header []byte, srcIP, dstIP string, headerLen uint8) uint16 {
	return checksum(append(append(append(header, net.ParseIP(srcIP)...), net.ParseIP(dstIP)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(headerLen >> 8), byte(headerLen)}...))
}
func verifyTCPchecksum() {
	// TODO: implement TCP checksum verification
}

func MyRawConnTCPWrite(w *ipv4.RawConn, tcp []byte, dst string) error {
	return w.WriteTo(&ipv4.Header{
		Version:  ipv4.Version, // protocol version
		Len:      IP_HEADER_LEN, // header length
		TOS:      0, // type-of-service (0 is everything normal)
		TotalLen: len(tcp) + IP_HEADER_LEN, // packet total length (octets)
		ID:       0, // identification
		Flags:    ipv4.DontFragment, // flags
		FragOff:  0, // fragment offset
		TTL:      DEFAULT_TTL, // time-to-live (maximum lifespan in seconds)
		Protocol: TCP_PROTO, // next protocol
		Checksum: 0, // checksum (apparently autocomputed)
		Dst: net.ParseIP(dst), // destination address
	}, tcp, nil)
}