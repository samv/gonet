package ping

import (
	"network/icmp"

	"github.com/hsheth2/logs"
)

func (pm *Ping_Manager) ping_replier() {
	for {
		ping := <-pm.input
		//logs.Info.Println("replying:", ping)
		go pm.respondTo(ping)
	}
}

func (pm *Ping_Manager) respondTo(ping *icmp.ICMP_In) error {
	ping.Header.TypeF = PING_ECHO_REPLY_TYPE

	// make packet
	err := ping.Header.MarshalICMPHeaderGivenSlice(ping.OriginalPacket)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// get writer
	writer, err := pm.getIP_Writer(ping.RIP)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// send
	//	logs.Info.Println("Send ping reply")
	err = writer.WriteTo(ping.OriginalPacket)
	if err != nil {
		logs.Error.Println(err)
	}

	return nil
}
