package arp

import (
	"network/ethernet"
	"github.com/hsheth2/logs"
	"errors"
)

type ARP_htype uint16
const ARP_HTYPE_ETHERNET = 1
const (
	ARP_OPER_REQUEST = 1
	ARP_OPER_REPLY = 2
)

type ARP_Manager struct {
	read chan *ethernet.Ethernet_Header
	ethtp_manager map[ethernet.EtherType](ARP_Protocol_Dealer)
}

var GlobalARP_Manager *ARP_Manager = func() *ARP_Manager {
	am, err := NewARP_Manager(ethernet.GlobalNetworkReader)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	return am
}()

func NewARP_Manager(in *ethernet.Network_Reader) (*ARP_Manager, error) {
	read, err := in.Bind(ethernet.ETHERTYPE_ARP)
	if err != nil {
		return nil, err
	}

	am := &ARP_Manager{
		read: read,
		ethtp_manager: make(map[ethernet.EtherType](ARP_Protocol_Dealer)),
	}

	go am.dealer()

	return am, nil
}

func (am *ARP_Manager) Register(tp ethernet.EtherType, arppd ARP_Protocol_Dealer) error {
	if tp == ethernet.ETHERTYPE_ARP {
		return errors.New("ARP Manager: cannot bind to ARP ethertype")
	}
	if _, ok := am.ethtp_manager[tp]; ok {
		return errors.New("ARP Manager: ethertype already bound to")
	}
	am.ethtp_manager[tp] = arppd
	return nil
}

// TODO make unregister function

func (am *ARP_Manager) dealer() {
	for {
		header := <-am.read
		data := header.Packet
		packet := ParseARP_Packet_General(data)
		if pd, ok := am.ethtp_manager[packet.ptype]; ok && packet.htype == ARP_HTYPE_ETHERNET {
			packet = ParseARP_Packet_Type(data, packet, pd)
			if packet.oper == ARP_OPER_REQUEST {
				// TODO reply to request
			} else if packet.oper == ARP_OPER_REPLY {
				// TODO deal with ARP reply
			} else {
				logs.Warn.Println("Dropping ARP packet for bad operation")
			}
		}
	}
}

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
		ptype: ethernet.EtherType(uint16(d[1]) << 8 | uint16(d[2])),
		hlen: uint8(d[3]),
		plen: uint8(d[4]),
		oper: uint16(d[5]) << 8 | uint16(d[6]),
	}
}

func ParseARP_Packet_Type(d []byte, packet *ARP_Packet, pd ARP_Protocol_Dealer) *ARP_Packet {
	packet.sha = &ethernet.MAC_Address{Data: d[7:7+packet.hlen]}
	packet.spa = pd.Unmarshal(d[7+packet.hlen:7+packet.hlen+packet.plen])
	packet.tha = &ethernet.MAC_Address{Data: d[7+packet.hlen+packet.plen:7+2*packet.hlen+packet.plen]}
	packet.tpa = pd.Unmarshal(d[7+2*packet.hlen+packet.plen:7+2*packet.hlen+2*packet.plen])
	return packet
}

type ARP_Protocol_Dealer interface {
	Lookup(ARP_Protocol_Address) (*ethernet.MAC_Address, error)
	Add(ARP_Protocol_Address, *ethernet.MAC_Address) error
	Unmarshal([]byte) ARP_Protocol_Address
}

type ARP_Protocol_Address interface {
	Marshal() ([]byte, error)
}
