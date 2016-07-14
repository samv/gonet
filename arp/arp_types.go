package arp

import (
	"github.com/hsheth2/gonet/ethernet"

	"github.com/hsheth2/notifiers"
)

// The ProtocolDealer is provided by a client during Registration
// for an EtherType. It provides this ARP package with all
// the EtherType specific address settings and functions.
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

// The ProtocolAddress represents a protocol address.
type ProtocolAddress interface {
	Marshal() ([]byte, error)
	Len() uint8
	ARPEqual(ProtocolAddress) bool
	String() string
}
