package icmpp

type ICMP_Header struct {
	typeF, code uint8
	cksum       uint16
	opt         uint32
	data        []byte
}

func (h *ICMP_Header) MarshalICMPHeader() ([]byte, error) {
	marshaledHeader := append([]byte{
		(byte)(h.typeF), (byte)(h.code), (byte)(h.cksum >> 8), (byte)(h.cksum),
	}, h.data...)
	return marshaledHeader, nil
}
