package icmp

import (
	"network/ipv4"
)

func SendICMPPacket(writer ipv4.Writer, data *Header) error {
	packet, err := data.Marshal()
	if err != nil {
		return err
	}

	_, err = writer.WriteTo(packet)
	return err
}
