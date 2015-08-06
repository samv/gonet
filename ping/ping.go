package ping

import (
	"network/icmpp"
	"network/ipv4p"
	"time"

	"github.com/hsheth2/logs"
)

const (
	PING_ECHO_REQUEST_TYPE = 8
	PING_ECHO_REPLY_TYPE   = 0
	PING_ICMP_CODE         = 0
)

type Ping_Manager struct {
	// Responding to pings
	input  chan *icmpp.ICMP_In
	output map[string](*ipv4p.IP_Writer)

	// Sending pings
	reply chan *icmpp.ICMP_In
}

func NewPing_Manager(icmprm *icmpp.ICMP_Read_Manager) (*Ping_Manager, error) {
	input, err := icmprm.Bind(8)
	if err != nil {
		return nil, err
	}

	reply, err := icmprm.Bind(0)
	if err != nil {
		return nil, err
	}

	pm := &Ping_Manager{
		input: input,
		output: make(map[string](*ipv4p.IP_Writer)),
		reply: reply,
	}

	go pm.ping_replier()

	return pm, nil
}

var GlobalPingManager = func() *Ping_Manager {
	pm, err := NewPing_Manager(icmpp.GlobalICMPReadManager)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return pm
}()

func (pm *Ping_Manager) getIP_Writer(ip string) (*ipv4p.IP_Writer, error) {
	if _, ok := pm.output[ip]; !ok {
		wt, err := ipv4p.NewIP_Writer(ip, ipv4p.ICMP_PROTO)
		if err != nil {
			return nil, err
		}
		pm.output[ip] = wt
	}
	return pm.output[ip], nil
}

func (pm *Ping_Manager) ping_replier() {
	for {
		ping := <-pm.input
		wr, err := pm.getIP_Writer(ping.RIP)
		if err != nil {
			logs.Error.Println(err)
			continue
		}
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

func (pm *Ping_Manager) SendPing(ip string, interval time.Duration) error {
	ipWriter, err := pm.getIP_Writer(ip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// prepare packet
	packet := &icmpp.ICMP_Header{
		TypeF: PING_ECHO_REQUEST_TYPE,
		Code:  PING_ICMP_CODE,
		Opt:   45<<16 | 1,
		Data:  []byte("abcdefg"),
	}

	// make data
	data, err := packet.MarshalICMPHeader()
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// send
	err = ipWriter.WriteTo(data)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	return nil
}
