package ping

import (
	"network/icmpp"
	"network/ipv4p"

	"github.com/hsheth2/logs"
)

func (pm *Ping_Manager) ping_replier() {
	for {
		ping := <-pm.input
		wr, err := pm.getIP_Writer(ping.RIP)
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//logs.Info.Println("replying:", ping)
		go pm.respondTo(wr, ping)
	}
}
func (pm *Ping_Manager) respondTo(writer *ipv4p.IP_Writer, ping *icmpp.ICMP_In) error {
	header := ping.Header
	header.TypeF = PING_ECHO_REPLY_TYPE

	// make packet
	bts, err := header.MarshalICMPHeader()
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// send
	err = writer.WriteTo(bts)
	if err != nil {
		logs.Error.Println(err)
	}

	return nil
}
