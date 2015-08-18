package ipv4tps

import (
	"net"
)

type Netmask uint8

type IPaddress string

const IPv4_ADDRESS_LENGTH = 4

func (ip *IPaddress) Marshal() ([]byte, error) {
	x := net.ParseIP(string(*ip))
	return x[12:], nil
}

func (ip *IPaddress) Len() uint8 {
	return IPv4_ADDRESS_LENGTH
}

func MakeIP(ip string) *IPaddress {
	p := IPaddress(ip)
	return &p
}
