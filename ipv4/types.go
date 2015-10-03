package ipv4

import (
	"encoding/binary"
	"net"
)

func initTypes() {
	allIP := []byte{0, 0, 0, 0}
	IPAll = &Address{IP: allIP}
	IPAllHash = IPAll.Hash()
}

// Address represents an IP address
type Address struct {
	IP []byte
}

// Netmask is a network's netmask
type Netmask uint8

// Hash is the type that is returned from an Address.Hash() call
type Hash uint32

// The length of an IPv4 Address
const IPv4AddressLength = 4

// Utilities for binding to all IP addresses
var (
	IPAll     *Address
	IPAllHash Hash
)

// Marshal turns an IP Address into a slice of bytes
func (ip *Address) Marshal() ([]byte, error) {
	return ip.IP, nil
}

// Hash converts an IP Address into a uint32 for hashing purposes
func (ip *Address) Hash() Hash {
	return Hash(binary.BigEndian.Uint32(ip.IP))
}

// Len returns the length of a marshaled IP address
func (ip *Address) Len() uint8 {
	return IPv4AddressLength
}

// MakeIP converts a string into an Address
func MakeIP(ip string) *Address {
	return &Address{
		IP: net.ParseIP(ip)[12:],
	}
}

func ipCompare(baseS, cmpS *Address, netm Netmask) bool {
	base := baseS.IP
	cmp := cmpS.IP

	// TODO take netmask into account
	for i := 0; i < len(base); i++ {
		if base[i] != cmp[i] && base[i] != 0 {
			return false
		}
	}
	return true
}
