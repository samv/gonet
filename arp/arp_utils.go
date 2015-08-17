package arp

import (
	"network/ethernet"
	"bytes"
)

type ARP_htype uint16
const ARP_HTYPE_ETHERNET = 1
const (
	ARP_OPER_REQUEST = 1
	ARP_OPER_REPLY = 2
)

type ARP_Packet struct {
	htype ARP_htype
	ptype ethernet.EtherType
	hlen, plen uint8
	oper uint16
	sha, tha *ethernet.MAC_Address
	spa, tpa ARP_Protocol_Address
}

func ParseARP_Packet_General(d []byte) *ARP_Packet {
	return &ARP_Packet{
		htype: ARP_htype(uint16(d[0]) << 8 | uint16(d[1])),
		ptype: ethernet.EtherType(uint16(d[2]) << 8 | uint16(d[3])),
		hlen: uint8(d[4]),
		plen: uint8(d[5]),
		oper: uint16(d[6]) << 8 | uint16(d[7]),
	}
}

func ParseARP_Packet_Type(d []byte, packet *ARP_Packet, pd ARP_Protocol_Dealer) *ARP_Packet {
	const offset = 8
	packet.sha = &ethernet.MAC_Address{Data: d[offset:offset+packet.hlen]}
	packet.spa = pd.Unmarshal(d[offset+packet.hlen:offset+packet.hlen+packet.plen])
	packet.tha = &ethernet.MAC_Address{Data: d[offset+packet.hlen+packet.plen:offset+2*packet.hlen+packet.plen]}
	packet.tpa = pd.Unmarshal(d[offset+2*packet.hlen+packet.plen:offset+2*packet.hlen+2*packet.plen])
	return packet
}

func (p *ARP_Packet) MarshalPacket() ([]byte, error) {
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