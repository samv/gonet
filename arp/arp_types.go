package arp

import (
	"network/ethernet"
	"github.com/hsheth2/notifiers"
)

type ProtocolDealer interface {
	Lookup(ProtocolAddress) (*ethernet.MACAddress, error)
	Request(ProtocolAddress) (*ethernet.MACAddress, error)
	// TODO add discover (probe) function to broadcast ARP requests
	// TODO support ARP announcements
	Add(ProtocolAddress, *ethernet.MACAddress) error
	GetReplyNotifier() *notifiers.Notifier
	Unmarshal([]byte) ProtocolAddress
	GetAddress() ProtocolAddress
}

type ProtocolAddress interface {
	Marshal() ([]byte, error)
	Len() uint8
	ARPEqual(ProtocolAddress) bool
}
