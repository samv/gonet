package icmp

import (
	"network/ipv4"
)

const ICMP_Header_MinSize = 8
const ICMP_QUEUE_Size = 10

type ICMP_Header struct {
	TypeF, Code uint8
	cksum       uint16
	Opt         uint32
	Data        []byte
}

type ICMP_In struct {
	Header   *ICMP_Header
	LIP, RIP ipv4.IPaddress
}

func (h *ICMP_Header) MarshalICMPHeader() ([]byte, error) {
	h.cksum = 0
	marshaledHeader := append([]byte{
		(byte)(h.TypeF), (byte)(h.Code), (byte)(h.cksum >> 8), (byte)(h.cksum),
		(byte)(h.Opt >> 24), (byte)(h.Opt >> 16), (byte)(h.Opt >> 8), (byte)(h.Opt),
	}, h.Data...)
	h.cksum = ipv4.Checksum(marshaledHeader)
	marshaledHeader[2] = byte(h.cksum >> 8)
	marshaledHeader[3] = byte(h.cksum)
	return marshaledHeader, nil
}

func ExtractICMPHeader(dat []byte, lip, rip ipv4.IPaddress) (*ICMP_In, error) {
	// TODO: ICMP checksum validation
	return &ICMP_In{
		Header: &ICMP_Header{
			TypeF: uint8(dat[0]),
			Code:  uint8(dat[1]),
			cksum: uint16(dat[2])<<8 | uint16(dat[3]),
			Opt:   uint32(dat[4])<<24 | uint32(dat[5])<<16 | uint32(dat[6])<<8 | uint32(dat[7]),
			Data:  dat[ICMP_Header_MinSize:],
		},
		LIP: lip,
		RIP: rip,
	}, nil
}
