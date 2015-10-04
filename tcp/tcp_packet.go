package tcp

import (
	"errors"
	"network/ipv4"
)

type packet struct {
	header   *header
	payload  []byte
	rip, lip *ipv4.Address
}

func (p *packet) Marshal() ([]byte, error) {
	head, err := p.header.Marshal(p.rip, p.lip, p.payload)
	packet := append(head, p.payload...)
	return packet, err
}

func (p *packet) getPayloadSize() uint32 {
	if len(p.payload) == 0 {
		return 1
	}
	return uint32(len(p.payload))
}

type header struct {
	srcport, dstport uint16
	seq, ack         uint32
	// will do data offset automatically
	flags  flag
	window uint16
	// checksum will be automatic
	urg     uint16
	options []byte
}

func (h *header) Marshal(dstIP, srcIP *ipv4.Address, data []byte) ([]byte, error) {
	// pad options with 0's
	for len(h.options)%4 != 0 {
		h.options = append(h.options, 0)
	}

	headerLen := uint16(basicHeaderSize + len(h.options)) // size of header in 32 bit (4 byte) chunks

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
	cksum := ipv4.CalcTransportChecksum(append(header, data...), srcIP, dstIP, headerLen+uint16(len(data)), ipv4.IPProtoTCP)
	header[16] = byte(cksum >> 8)
	header[17] = byte(cksum)

	return header, nil
}

func extractPacket(d []byte, rip, lip *ipv4.Address) (*packet, error) {

	// header length
	headerLen := uint16((d[12] >> 4) * 4)
	if headerLen < basicHeaderSize {
		return nil, errors.New("Bad TCP header size: Less than 20.")
	}

	// checksum verification
	if !ipv4.VerifyTransportChecksum(d[:headerLen], rip, lip, headerLen, ipv4.IPProtoTCP) {
		return nil, errors.New("Bad TCP header checksum")
	}

	// create the header
	h := &header{
		srcport: uint16(d[0])<<8 | uint16(d[1]),
		dstport: uint16(d[2])<<8 | uint16(d[3]),
		seq:     uint32(d[4])<<24 | uint32(d[5])<<16 | uint32(d[6])<<8 | uint32(d[7]),
		ack:     uint32(d[8])<<24 | uint32(d[9])<<16 | uint32(d[10])<<8 | uint32(d[11]),
		flags:   flag(d[13]),
		window:  uint16(d[14])<<8 | uint16(d[15]),
		urg:     uint16(d[18])<<8 | uint16(d[19]),
		options: d[basicHeaderSize:headerLen],
	}
	return &packet{header: h, payload: d[headerLen:], rip: rip, lip: lip}, nil
}
