package icmp

import (
	"network/ipv4"
	"network/ipv4/ipv4tps"
)

const ICMP_Header_MinSize = 8
const ICMP_QUEUE_Size = 100

type ICMP_Header struct {
	TypeF, Code uint8
	cksum       uint16
	Opt         uint32
	Data        []byte
}

type ICMP_In struct {
	Header         *ICMP_Header
	OriginalPacket []byte
	LIP, RIP       ipv4tps.IPaddress
}

func (h *ICMP_Header) MarshalICMPHeader() ([]byte, error) {
	base := make([]byte, ICMP_Header_MinSize+len(h.Data))
	copy(base[ICMP_Header_MinSize:], h.Data)
	err := h.MarshalICMPHeaderGivenSlice(base)
	return base, err
}

func (h *ICMP_Header) MarshalICMPHeaderGivenSlice(base []byte) error {
	base[0] = (byte)(h.TypeF)
	base[1] = (byte)(h.Code)
	base[2] = 0 // checksum; set later
	base[3] = 0
	base[4] = (byte)(h.Opt >> 24)
	base[5] = (byte)(h.Opt >> 16)
	base[6] = (byte)(h.Opt >> 8)
	base[7] = (byte)(h.Opt)
	h.cksum = ipv4.Checksum(base)
	base[2] = byte(h.cksum >> 8)
	base[3] = byte(h.cksum)
	return nil
}

func ExtractICMPHeader(dat []byte, lip, rip *ipv4tps.IPaddress) (*ICMP_In, error) {
	// TODO: ICMP checksum validation
	return &ICMP_In{
		Header: &ICMP_Header{
			TypeF: uint8(dat[0]),
			Code:  uint8(dat[1]),
			cksum: uint16(dat[2])<<8 | uint16(dat[3]),
			Opt:   uint32(dat[4])<<24 | uint32(dat[5])<<16 | uint32(dat[6])<<8 | uint32(dat[7]),
			Data:  dat[ICMP_Header_MinSize:],
		},
		LIP:            lip,
		RIP:            rip,
		OriginalPacket: dat,
	}, nil
}
