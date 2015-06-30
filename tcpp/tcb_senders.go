package tcpp

import (
	"errors"
	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
	"golang.org/x/net/ipv4"
	"net"
	"network/ipv4p"
	"sync"
	"time"
)

func (c *TCB) UpdateState(newState uint) {
	logs.Info.Println("The New State is", newState)
	c.state = newState
	go SendUpdate(c.stateUpdate)
	if c.serverParent != nil {
		go SendUpdate(c.serverParent.connQueueUpdate)
	}
}

func (c *TCB) UpdateLastAck(newAck uint32) error {
	logs.Info.Println("New ack number:", newAck)
	c.recentAckNum = newAck
	go notifiers.SendNotifierBroadcast(c.recentAckUpdate, c.recentAckNum)
	return nil
}

func SendUpdate(update *sync.Cond) {
	update.L.Lock()
	update.Broadcast()
	update.L.Unlock()
}

func (c *TCB) PacketSender() {
	// TODO: deal with data in send and urgSend buffers
}

func (c *TCB) SendWithRetransmit(data *TCP_Packet) error {
	// send the first packet
	c.SendPacket(data)

	// ack listeners
	ackFound := make(chan bool, 1)
	killAckListen := make(chan bool, 1)
	go c.ListenForAck(ackFound, killAckListen, data.header.seq+data.getPayloadSize())

	// timers and timeouts
	resendTimer := make(chan bool, TCP_RESEND_LIMIT)
	timeout := make(chan bool, 1)
	killTimer := make(chan bool, 1)
	go ResendTimer(resendTimer, timeout, killTimer, c.resendDelay)

	// resend if needed
	for {
		select {
		case <-ackFound:
			killTimer <- true
			return nil
		case <-resendTimer:
			c.SendPacket(data)
		case <-timeout:
			// TODO deal with a resend timeout fully
			killAckListen <- true
			logs.Error.Println("Resend of packet seq", data.header.seq, "timed out")
			return errors.New("Resend timed out")
		}
	}
}

func (c *TCB) ListenForAck(successOut chan<- bool, end <-chan bool, targetAck uint32) {
	logs.Trace.Println("Listening for ack:", targetAck)
	in := c.recentAckUpdate.Register(ACK_BUF_SZ)
	defer c.recentAckUpdate.Unregister(in)
	for {
		select {
		case v := <-in:
			logs.Trace.Println("Ack listener got ack: ", v.(uint32))
			if v.(uint32) == targetAck {
				logs.Trace.Println("Killing the resender for ack:", v.(uint32))
				successOut <- true
				return
			}
		case <-end:
			return
		}
	}
}

func ResendTimer(timerOutput, timeout chan<- bool, finished <-chan bool, delay time.Duration) {
	for i := 0; i < TCP_RESEND_LIMIT; i++ {
		select {
		case <-time.After(delay):
			timerOutput <- true
			delay *= 2 // increase the delay after each resend
		case <-finished:
			return
		}
	}
	timeout <- true
}

func (c *TCB) SendPacket(d *TCP_Packet) error {
	// Requires that seq, ack, flags, urg, and options are set
	// Will set everything else

	d.header.srcport = c.lport
	d.header.dstport = c.rport
	d.header.window = c.curWindow // TODO improve the window field calculation
	d.rip = c.ipAddress
	d.lip = c.srcIP

	pay, err := d.Marshal_TCP_Packet()
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	err = c.writer.WriteTo(&ipv4.Header{
		Version:  ipv4.Version,                   // protocol version
		Len:      ipv4p.IP_HEADER_LEN,            // header length
		TOS:      0,                              // type-of-service (0 is everything normal)
		TotalLen: len(pay) + ipv4p.IP_HEADER_LEN, // packet total length (octets)
		ID:       0,                              // identification
		Flags:    ipv4.DontFragment,              // flags
		FragOff:  0,                              // fragment offset
		TTL:      ipv4p.DEFAULT_TTL,              // time-to-live (maximum lifespan in seconds)
		Protocol: ipv4p.TCP_PROTO,                // next protocol
		Checksum: 0,                              // checksum (autocomputed)
		Dst:      net.ParseIP(d.rip),             // destination address
	}, pay, nil)

	if err != nil {
		logs.Error.Println(err)
		return err
	}

	return nil
}

func (c *TCB) SendReset(seq uint32, ack uint32) error {
	logs.Trace.Println("Sending RST with seq: ", seq, " and ack: ", ack)
	rst := &TCP_Packet{
		header: &TCP_Header{
			seq:     seq,
			ack:     ack,
			flags:   TCP_RST,
			urg:     0,
			options: []byte{},
		},
		payload: []byte{},
	}

	return c.SendPacket(rst)
}

func (c *TCB) SendAck(seq, ack uint32) error {
	logs.Trace.Println("Sending ACK with seq: ", seq, " and ack: ", ack)
	ack_packet := &TCP_Packet{
		header: &TCP_Header{
			seq:     seq,
			ack:     ack,
			flags:   TCP_ACK,
			urg:     0,
			options: []byte{},
		},
		payload: []byte{},
	}
	return c.SendPacket(ack_packet)
}
