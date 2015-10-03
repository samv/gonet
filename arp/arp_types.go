package arp

import (
	"network/ethernet"
	"github.com/hsheth2/notifiers"
)

type ARP_Protocol_Dealer interface {
	Lookup(ARP_Protocol_Address) (*ethernet.MACAddress, error)
	Request(ARP_Protocol_Address) (*ethernet.MACAddress, error)
	// TODO add discover (probe) function to broadcast ARP requests
	// TODO support ARP announcements
	Add(ARP_Protocol_Address, *ethernet.MACAddress) error
	GetReplyNotifier() *notifiers.Notifier
	Unmarshal([]byte) ARP_Protocol_Address
	GetAddress() ARP_Protocol_Address
}

type ARP_Protocol_Address interface {
	Marshal() ([]byte, error)
	Len() uint8
	ARPEqual(ARP_Protocol_Address) bool
}
