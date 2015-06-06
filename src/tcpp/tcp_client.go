package tcpp

import (
	"errors"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
	"logs"
	"ipv4p"
)

func New_TCB_From_Client(local, remote uint16, dstIP string) (*TCB, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	read, err := TCP_Port_Manager.bind(remote, local, dstIP)
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", ipv4p.TCP_PROTO), dstIP) // only for read, not for write
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	r, err := ipv4.NewRawConn(p)
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	logs.Trace.Println("Finished New TCB from Client")
	return New_TCB(local, remote, dstIP, read, r, TCP_CLIENT)
}

func (c *TCB) Connect() error {
	if c.kind != TCP_CLIENT || c.state != CLOSED {
		return errors.New("TCB is not a closed client")
	}

	// Build the SYN packet
	SYN := &TCP_Packet{
		header: &TCP_Header{
			srcport: c.lport,
			dstport: c.rport,
			seq:     c.seqNum,
			ack:     c.ackNum,
			flags:   TCP_SYN,
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
	logs.Trace.Println("About to send syn")
	go c.SendWithRetransmit(SYN)
	logs.Trace.Println("Sent SYN")
	c.UpdateState(SYN_SENT)

	// wait for the connection to be established
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	for c.state != ESTABLISHED && c.state != CLOSED {
		c.stateUpdate.Wait()
	}
	if c.state == CLOSED {
		return errors.New("Connection closed by reset or timeout")
	} else {
		return nil
	}
}
