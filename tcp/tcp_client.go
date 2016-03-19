package tcp

import (
	"errors"
	"network/ipv4"

	"github.com/hsheth2/logs"
)

type Client struct {
	tcb *TCB
}

func NewClient(remote uint16, dstIP *ipv4.Address) (*Client, error) {
	local, err := portManager.GetUnusedPort()
	if err != nil {
		logs.Error.Println(err, remote, dstIP.Hash())
		return nil, err
	}

	read, err := portManager.bind(remote, local, dstIP)
	if err != nil {
		logs.Error.Println(err, remote, dstIP.Hash())
		return nil, err
	}

	r, err := ipv4.NewWriter(dstIP, ipv4.IPProtoTCP)
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	/*logs*/logs.Trace.Println("Finished New TCB from Client")
	tcb, err := newTCB(local, remote, dstIP, read, r, clientParent)
	if err != nil {
		return nil, err
	}

	return &Client{tcb: tcb}, nil
}

func (c *Client) Connect() (*TCB, error) {
	if c.tcb.kind != clientParent || c.tcb.getState() != fsmClosed {
		return nil, errors.New("not a closed client")
	}

	// Build the SYN packet
	SYN := &packet{
		header: &header{
			srcport: c.tcb.lport,
			dstport: c.tcb.rport,
			seq:     c.tcb.seqNum,
			ack:     c.tcb.ackNum,
			flags:   flagSyn,
			window:  c.tcb.curWindow, // TODO improve the window size calculation
			urg:     0,
			options: []byte{0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0x0,
				0x0, 0x0, 0x0, 0x01, 0x03, 0x03, 0x07}, // TODO compute the options of SYN instead of hardcoding them
		},
		payload: []byte{},
		rip:     c.tcb.ipAddress,
		lip:     c.tcb.srcIP,
	}
	c.tcb.seqNum += 1

	// Send the SYN packet
	/*logs*/logs.Trace.Println(c.tcb.hash(), "About to send syn")
	c.tcb.updateState(fsmSynSent)
	go c.tcb.sendWithRetransmit(SYN)
	/*logs*/logs.Trace.Println(c.tcb.hash(), "Sent SYN")

	// wait for the connection to be established
	c.tcb.stateUpdate.L.Lock()
	defer c.tcb.stateUpdate.L.Unlock()
	for c.tcb.state != fsmEstablished && c.tcb.state != fsmClosed {
		c.tcb.stateUpdate.Wait()
	}
	if c.tcb.state == fsmClosed {
		return nil, errors.New("connection closed by reset or timeout")
	} else {
		return c.tcb, nil
	}
}
