package ipv4

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

// String marshals an IP address into dotted integer form with optional 0's removed.
func (ip *Address) String() string {
	//if ip.IP[2] == 0 {
	//if ip.IP[1] == 0 {
	//return fmt.Sprintf("%d.%d", ip.IP[0], ip.IP[3])
	//} else {
	//return fmt.Sprintf("%d.%d.%d", ip.IP[0], ip.IP[1], ip.IP[3])
	//}
	//} else {
	return fmt.Sprintf("%d.%d.%d.%d", ip.IP[0], ip.IP[1], ip.IP[2], ip.IP[3])
	//}
}

// Hash converts an IP Address into a uint32 for hashing purposes
func (ip *Address) Hash() Hash {
	return Hash(binary.BigEndian.Uint32(ip.IP))
}

// Len returns the length of a marshaled IP address
func (ip *Address) Len() uint8 {
	return IPv4AddressLength
}

func (ip *Address) Equal(other *Address) bool {
	return bytes.Equal(ip.IP, other.IP)
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
