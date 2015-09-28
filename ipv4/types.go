package ipv4

import (
	"encoding/binary"
	"net"
)

func initTypes() {
	allIP := []byte{0, 0, 0, 0}
	IPAll = &IPAddress{IP: allIP}
	IPAllHash = IPAll.Hash()
}

type Netmask uint8

type IPhash uint32
type IPAddress struct {
	IP []byte
}

const IPv4AddressLength = 4

var (
	IPAll     *IPAddress
	IPAllHash IPhash
)

func (ip *IPAddress) Marshal() ([]byte, error) {
	return ip.IP, nil
}

func (ip *IPAddress) Hash() IPhash {
	return IPhash(binary.BigEndian.Uint32(ip.IP))
}

func (ip *IPAddress) Len() uint8 {
	return IPv4AddressLength
}

func MakeIP(ip string) *IPAddress {
	return &IPAddress{
		IP: net.ParseIP(ip)[12:],
	}
}
