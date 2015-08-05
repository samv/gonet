package icmpp

import "network/ipv4p"

type ICMP_Header struct {
	typeF, code uint8
	cksum       uint16
	opt         uint32
	data        []byte
}

func (h *ICMP_Header) MarshalICMPHeader() ([]byte, error) {
	h.cksum = 0
	marshaledHeader := append([]byte{
		(byte)(h.typeF), (byte)(h.code), (byte)(h.cksum >> 8), (byte)(h.cksum), (byte)(h.opt >> 24), (byte)(h.opt >> 16), (byte)(h.opt >> 8), (byte)(h.opt),
	}, h.data...)
	h.cksum = ipv4p.Checksum(marshaledHeader)
	marshaledHeader[2] = byte(h.cksum >> 8)
	marshaledHeader[3] = byte(h.cksum)
	return marshaledHeader, nil
}
