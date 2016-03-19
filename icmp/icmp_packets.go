package icmp

import (
	"github.com/hsheth2/gonet/ipv4"
)

// Header represents an ICMP header
type Header struct {
	Tp    Type
	Code  uint8
	Opt   uint32
	Data  []byte
	cksum uint16
}

// Packet contains an ICMP header and some information from the
// internet layer protocol header
type Packet struct {
	Header         *Header
	OriginalPacket []byte
	LIP, RIP       *ipv4.Address
}

// Marshal converts a Header structure into a slice of bytes,
// ready for transmission.
func (h *Header) Marshal() ([]byte, error) {
	base := make([]byte, HeaderMinSize+len(h.Data))
	copy(base[HeaderMinSize:], h.Data)
	err := h.MarshalGivenSlice(base)
	return base, err
}

// MarshalGivenSlice populates a given slice of bytes
// with fields from the Header structure.
func (h *Header) MarshalGivenSlice(base []byte) error {
	// basics
	base[0] = (byte)(h.Tp)
	base[1] = (byte)(h.Code)
	base[2] = 0 // checksum; set at end
	base[3] = 0
	base[4] = (byte)(h.Opt >> 24)
	base[5] = (byte)(h.Opt >> 16)
	base[6] = (byte)(h.Opt >> 8)
	base[7] = (byte)(h.Opt)

	// checksum
	h.cksum = ipv4.Checksum(base)
	base[2] = byte(h.cksum >> 8)
	base[3] = byte(h.cksum)
	return nil
}

func extractHeader(dat []byte, lip, rip *ipv4.Address) (*Packet, error) {
	// TODO: ICMP checksum validation
	return &Packet{
		Header: &Header{
			Tp:    Type(dat[0]),
			Code:  uint8(dat[1]),
			cksum: uint16(dat[2])<<8 | uint16(dat[3]),
			Opt:   uint32(dat[4])<<24 | uint32(dat[5])<<16 | uint32(dat[6])<<8 | uint32(dat[7]),
			Data:  dat[HeaderMinSize:],
		},
		LIP:            lip,
		RIP:            rip,
		OriginalPacket: dat,
	}, nil
}
