package tcp

import (
	"errors"
	"network/ipv4"

	"github.com/hsheth2/logs"
)

func New_TCB_From_Client(local, remote uint16, dstIP *ipv4.Address) (*TCB, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	read, err := TCP_Port_Manager.bind(remote, local, dstIP)
	if err != nil {
		logs.Error.Println(err, local, remote, dstIP.Hash())
		return nil, err
	}

	r, err := ipv4.NewWriter(dstIP, ipv4.IPProtoTCP)
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	//ch logs.Trace.Println("Finished New TCB from Client")
	return New_TCB(local, remote, dstIP, read, r, clientParent)
}

func (c *TCB) Connect() error {
	if c.kind != clientParent || c.getState() != fsmClosed {
		return errors.New("TCB is not a closed client")
	}

	// Build the SYN packet
	SYN := &packet{
		header: &header{
			srcport: c.lport,
			dstport: c.rport,
			seq:     c.seqNum,
			ack:     c.ackNum,
			flags:   flagSyn,
			window:  c.curWindow, // TODO improve the window size calculation
			urg:     0,
			options: []byte{0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0x0, 0x0, 0x0, 0x0, 0x01, 0x03, 0x03, 0x07}, // TODO compute the options of SYN instead of hardcoding them
		},
		payload: []byte{},
		rip:     c.ipAddress,
		lip:     c.srcIP,
	}
	c.seqNum += 1

	// Send the SYN packet
	//ch logs.Trace.Println(c.Hash(), "About to send syn")
	c.updateState(fsmSynSent)
	go c.sendWithRetransmit(SYN)
	//ch logs.Trace.Println(c.Hash(), "Sent SYN")

	// wait for the connection to be established
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	for c.state != fsmEstablished && c.state != fsmClosed {
		c.stateUpdate.Wait()
	}
	if c.state == fsmClosed {
		return errors.New("Connection closed by reset or timeout")
	} else {
		return nil
	}
}
