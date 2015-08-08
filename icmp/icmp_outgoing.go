package icmp

import (
	"network/ipv4"
)

func SendICMPPacket(writer *ipv4.IP_Writer, data *ICMP_Header) error {
	packet, err := data.MarshalICMPHeader()
	if err != nil {
		return err
	}

	return writer.WriteTo(packet)
}
