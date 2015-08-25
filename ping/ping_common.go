package ping

import (
	"network/icmp"
	"network/ipv4"

	"network/ipv4/ipv4tps"

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
	input  chan *icmp.ICMP_In
	output map[ipv4tps.IPhash](*ipv4.IP_Writer)

	// Sending pings and receiving responses
	reply             chan *icmp.ICMP_In
	currentIdentifier uint16
	identifiers       map[uint16](chan *icmp.ICMP_In)
}

func NewPing_Manager(icmprm *icmp.ICMP_Read_Manager) (*Ping_Manager, error) {
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
		output:            make(map[ipv4tps.IPhash](*ipv4.IP_Writer), 1),
		reply:             reply,
		currentIdentifier: PING_START_ID,
		identifiers:       make(map[uint16](chan *icmp.ICMP_In)),
	}

	go pm.ping_replier()
	go pm.ping_response_dealer()

	return pm, nil
}

var GlobalPingManager = func() *Ping_Manager {
	pm, err := NewPing_Manager(icmp.GlobalICMPReadManager)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return pm
}()

func (pm *Ping_Manager) getIP_Writer(ip *ipv4tps.IPaddress) (*ipv4.IP_Writer, error) {
	if x, ok := pm.output[ip.Hash()]; ok {
		return x, nil
	}
	wt, err := ipv4.NewIP_Writer(ip, ipv4.ICMP_PROTO)
	if err != nil {
		return nil, err
	}
	pm.output[ip.Hash()] = wt
	return wt, nil
}
