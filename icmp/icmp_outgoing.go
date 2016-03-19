package icmp

import (
	"github.com/hsheth2/gonet/ipv4"
)

// SendPacket sends an ICMP packet to on a given writer
func SendPacket(writer ipv4.Writer, data *Header) error {
	packet, err := data.Marshal()
	if err != nil {
		return err
	}

	_, err = writer.WriteTo(packet)
	return err
}
