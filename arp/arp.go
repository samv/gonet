package arp

import (
	"network/ethernet"
	"github.com/hsheth2/logs"
	"errors"
)

type ARP_Manager struct {
	read chan *ethernet.Ethernet_Header
	write *ethernet.Network_Writer
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

	write, err := ethernet.NewNetwork_Writer()
	if err != nil {
		return nil, err
	}

	am := &ARP_Manager{
		read: read,
		write: write,
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
//			logs.Trace.Println("ARP packet:", packet)
			if packet.oper == ARP_OPER_REQUEST {
//				logs.Trace.Println("Got ARP Request")
				reply := &ARP_Packet{
					htype: packet.htype,
					ptype: packet.ptype,
					hlen: packet.hlen,
					plen: packet.plen,
					oper: ARP_OPER_REPLY,
					sha: ethernet.External_mac_address,
					spa: pd.GetAddress(),
					tha: packet.sha,
					tpa: packet.spa,
				}
				rp, err := reply.MarshalPacket()
				if err != nil {
					logs.Warn.Println("MarshalPacket failed; dropping ARP request")
					continue
				}
				am.write.Write(rp, reply.tha, ethernet.ETHERTYPE_ARP)
				logs.Trace.Println("Replied to ARP request")
			} else if packet.oper == ARP_OPER_REPLY {
				logs.Trace.Println("Got ARP Reply")
				// TODO deal with ARP reply
			} else {
				logs.Warn.Println("Dropping ARP packet for bad operation")
			}
		}
	}
}

type ARP_Protocol_Dealer interface {
	Lookup(ARP_Protocol_Address) (*ethernet.MAC_Address, error)
	Add(ARP_Protocol_Address, *ethernet.MAC_Address) error
	Unmarshal([]byte) ARP_Protocol_Address
	GetAddress() ARP_Protocol_Address
}

type ARP_Protocol_Address interface {
	Marshal() ([]byte, error)
}
