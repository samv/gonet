package ping

import (
	"network/icmpp"
	"network/ipv4p"

	"github.com/hsheth2/logs"
)

const (
	PING_ECHO_REQUEST_TYPE = 8
	PING_ECHO_REPLY_TYPE   = 0
	PING_ICMP_CODE         = 0
	PING_START_ID          = 8000 // TODO choose this randomly
)

type Ping_Manager struct {
	// Responding to pings
	input  chan *icmpp.ICMP_In
	output map[string](*ipv4p.IP_Writer)

	// Sending pings and receiving responses
	reply             chan *icmpp.ICMP_In
	currentIdentifier uint16
	identifiers       map[uint16](chan *icmpp.ICMP_In)
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
		input:             input,
		output:            make(map[string](*ipv4p.IP_Writer)),
		reply:             reply,
		currentIdentifier: PING_START_ID,
		identifiers:       make(map[uint16](chan *icmpp.ICMP_In)),
	}

	go pm.ping_replier()
	go pm.ping_response_dealer()

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
