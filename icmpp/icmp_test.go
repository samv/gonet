package icmpp

import (
	"fmt"
	"network/ipv4p"
	"testing"
)

func TestICMP(t *testing.T) {
	ipWriter, err := ipv4p.NewIP_Writer("127.0.0.1", 1)
	if err != nil {
		fmt.Println("error making ip writer")
		return
	}
	packet, err := ICMP_Header{
		typeF: 8,
		code:  0,
		opt:   nil,
	}.MarshalICMPHeader()
	if err != nil {
		fmt.Println("Error marshaling icmp header")
		return
	}
	ipWriter.WriteTo(packet)
}
