package ping

import (
	"network/icmp"
	"network/ipv4"

	"github.com/hsheth2/logs"
)

const (
	PING_ICMP_CODE = 0
	PING_START_ID  = 8000 // TODO choose this randomly
)

type Ping_Manager struct {
	// Responding to pings
	input  chan *icmp.Packet
	output map[ipv4.Hash](ipv4.Writer)

	// Sending pings and receiving responses
	reply             chan *icmp.Packet
	currentIdentifier uint16
	identifiers       map[uint16](chan *icmp.Packet)
}

func NewPing_Manager() (*Ping_Manager, error) {
	input, err := icmp.Bind(icmp.EchoRequest)
	if err != nil {
		return nil, err
	}

	reply, err := icmp.Bind(icmp.EchoReply)
	if err != nil {
		return nil, err
	}

	pm := &Ping_Manager{
		input:             input,
		output:            make(map[ipv4.Hash](ipv4.Writer), 1),
		reply:             reply,
		currentIdentifier: PING_START_ID,
		identifiers:       make(map[uint16](chan *icmp.Packet)),
	}

	go pm.ping_replier()
	go pm.ping_response_dealer()

	return pm, nil
}

var GlobalPingManager = func() *Ping_Manager {
	pm, err := NewPing_Manager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return pm
}()

func (pm *Ping_Manager) getIP_Writer(ip *ipv4.Address) (ipv4.Writer, error) {
	if x, ok := pm.output[ip.Hash()]; ok {
		return x, nil
	}
	wt, err := ipv4.NewWriter(ip, ipv4.IPProtoICMP)
	if err != nil {
		return nil, err
	}
	pm.output[ip.Hash()] = wt
	return wt, nil
}
