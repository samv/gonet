package icmp

import (
	"network/ipv4"
)

func SendICMPPacket(writer ipv4.Writer, data *ICMP_Header) error {
	packet, err := data.MarshalICMPHeader()
	if err != nil {
		return err
	}

	_, err = writer.WriteTo(packet)
	return err
}
