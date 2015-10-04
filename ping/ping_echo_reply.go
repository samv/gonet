package ping

import (
	"network/icmp"

	"github.com/hsheth2/logs"
)

func (pm *Ping_Manager) ping_replier() {
	for {
		ping := <-pm.input
		////ch logs.Info.Println("replying:", ping)
		go pm.respondTo(ping)
	}
}

func (pm *Ping_Manager) respondTo(ping *icmp.Packet) error {
	ping.Header.Tp = icmp.EchoReply

	// get writer
	writer, err := pm.getIP_Writer(ping.RIP)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// send
	//	//ch logs.Info.Println("Send ping reply")
	err = icmp.SendICMPPacket(writer, ping.Header)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	return nil
}
