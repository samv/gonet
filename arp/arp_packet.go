package arp

import (
	"bytes"

	"github.com/hsheth2/gonet/ethernet"
)

type packet struct {
	htype      hType
	ptype      ethernet.EtherType
	hlen, plen len
	oper       uint16
	sha, tha   *ethernet.MACAddress
	spa, tpa   ProtocolAddress
}

func parsePacket(d []byte) *packet {
	return &packet{
		htype: hType(uint16(d[0])<<8 | uint16(d[1])),
		ptype: ethernet.EtherType(uint16(d[2])<<8 | uint16(d[3])),
		hlen:  len(d[4]),
		plen:  len(d[5]),
		oper:  uint16(d[6])<<8 | uint16(d[7]),
	}
}

func parsePacketWithType(d []byte, packet *packet, pd ProtocolDealer) *packet {
	const offset = 8
	packet.sha = &ethernet.MACAddress{Data: d[offset : offset+packet.hlen]}
	packet.spa = pd.Unmarshal(d[offset+packet.hlen : offset+packet.hlen+packet.plen])
	packet.tha = &ethernet.MACAddress{Data: d[offset+packet.hlen+packet.plen : offset+2*packet.hlen+packet.plen]}
	packet.tpa = pd.Unmarshal(d[offset+2*packet.hlen+packet.plen : offset+2*packet.hlen+2*packet.plen])
	return packet
}

func (p *packet) MarshalPacket() ([]byte, error) {
	head := []byte{
		byte(p.htype >> 8), byte(p.htype),
		byte(p.ptype >> 8), byte(p.ptype),
		byte(p.hlen),
		byte(p.plen),
		byte(p.oper >> 8), byte(p.oper),
	}
	spa, err := p.spa.Marshal()
	if err != nil {
		return nil, err
	}
	tpa, err := p.tpa.Marshal()
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{head, p.sha.Data, spa, p.tha.Data, tpa}, nil), nil
}
