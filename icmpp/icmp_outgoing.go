package icmpp

import (
	"network/ipv4p"
)

func SendICMPPacket(writer *ipv4p.IP_Writer, data *ICMP_Header) error {
	packet, err := data.MarshalICMPHeader()
	if err != nil {
		return err
	}

	return writer.WriteTo(packet)
}
