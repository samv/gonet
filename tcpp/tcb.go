package tcpp

import (
	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
	"golang.org/x/net/ipv4"
	"sync"
	"time"
	"network/ipv4p"
)

type TCB struct {
	read            chan *TCP_Packet    // input
	writer          *ipv4.RawConn       // output
	ipAddress       string              // destination ip address
	srcIP           string              // src ip address
	lport, rport    uint16              // ports
	seqNum          uint32              // seq number (SND.NXT)
	ackNum          uint32              // ack number (RCV.NXT)
	state           uint                // from the FSM
	stateUpdate     *sync.Cond          // signals when the state is changed
	kind            uint                // type (server or client)
	serverParent    *Server_TCB         // the parent server
	curWindow       uint16              // the current window size
	sendBuffer      []byte              // a buffer of bytes that need to be sent
	urgSendBuffer   []byte              // buffer of urgent data TODO urg data later
	recvBuffer      []byte              // bytes to pass to the application above
	resendDelay     time.Duration       // the delay before resending
	ISS             uint32              // the initial snd seq number
	IRS             uint32              // the initial rcv seq number
	recentAckNum    uint32              // the last ack received (also SND.UNA)
	recentAckUpdate *notifiers.Notifier // signals changes in recentAckNum
}

func New_TCB(local, remote uint16, dstIP string, read chan *TCP_Packet, write *ipv4.RawConn, kind uint) (*TCB, error) {
	logs.Trace.Println("New_TCB")

	seq, err := genRandSeqNum()
	if err != nil {
		logs.Error.Fatal(err)
		return nil, err
	}

	c := &TCB{
		lport:           local,
		rport:           remote,
		ipAddress:       dstIP,
		srcIP:           ipv4p.GetSrcIP(dstIP),
		read:            read,
		writer:          write,
		seqNum:          seq,
		ackNum:          uint32(0), // Always 0 at start
		state:           CLOSED,
		stateUpdate:     sync.NewCond(&sync.Mutex{}),
		kind:            kind,
		serverParent:    nil,
		curWindow:       43690, // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
		resendDelay:     250 * time.Millisecond,
		ISS:             seq,
		IRS:             0,
		recentAckNum:    0,
		recentAckUpdate: notifiers.NewNotifier(),
	}
	logs.Trace.Println("Starting the packet dealer")

	go c.PacketSender()
	go c.PacketDealer()

	return c, nil
}

func (c *TCB) PacketDealer() {
	// read each tcp packet and deal with it
	logs.Trace.Println("Packet Dealing")
	for {
		logs.Trace.Println("Waiting for packets")
		segment := <-c.read
		logs.Trace.Println("got a packet")

		// First check if closed, listen, or syn-sent state
		switch c.state {
		case CLOSED:
			logs.Trace.Println("Dealing closed")
			c.DealClosed(segment)
			continue
		case LISTEN:
			logs.Trace.Println("Dealing listen")
			c.DealListen(segment)
			continue
		case SYN_SENT:
			c.DealSynSent(segment)
			continue
		}

		// TODO check sequence number

		if segment.header.flags&TCP_RST != 0 {
			// TODO finish: page 70
			switch c.state {
			case SYN_RCVD:
				// TODO not done
				continue
			case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2, CLOSE_WAIT:
				// TODO not done
				continue
			case CLOSING, LAST_ACK, TIME_WAIT:
				if segment.header.flags&TCP_RST != 0 {
					c.UpdateState(CLOSED)
				}
				continue
			}
		}

		// TODO check security/precedence
		// TODO check SYN (SYN bit shouldn't be there)

		if segment.header.flags&TCP_ACK == 0 {
			logs.Info.Println("Dropping a packet without an ACK flag")
			continue
		}

		switch c.state {
		case SYN_RCVD:
			if c.recentAckNum <= segment.header.ack && segment.header.ack <= c.seqNum {
				c.UpdateState(ESTABLISHED)
			} else {
				err := c.SendReset(segment.header.ack, 0)
				logs.Info.Println("Sent RST data")
				if err != nil {
					logs.Error.Println(err)
					continue
				}
			}
		case ESTABLISHED:
			if c.recentAckNum < segment.header.ack && segment.header.ack <= c.seqNum {
				c.UpdateLastAck(segment.header.ack)
				// TODO handle retrans queue
				// TODO update send window
			} else if c.recentAckNum > segment.header.ack {
				// ignore
				logs.Info.Println("Dropping packet: ACK validation failed")
				continue
			} else if segment.header.ack > c.seqNum {
				// TODO send ack, drop segment, return
				logs.Info.Println("Dropping packet with bad ACK field")
				continue
			}
		case FIN_WAIT_1:
			// TODO if acking fin
			c.UpdateState(FIN_WAIT_2)
		case FIN_WAIT_2:
			// TODO if retrans queue empty, acknowledge user's close with ok
		case CLOSE_WAIT:
			if c.recentAckNum < segment.header.ack && segment.header.ack <= c.seqNum {
				c.recentAckNum = segment.header.ack
				// TODO handle retrans queue
				// TODO update send window
			} else if c.recentAckNum > segment.header.ack {
				// ignore
				continue
			} else if segment.header.ack > c.seqNum {
				// TODO send ack, drop segment, return
			}
		case CLOSING:
			// TODO if ack is acknowledging our fin
			c.UpdateState(TIME_WAIT)
			// TODO else drop segment
		case LAST_ACK:
			// TODO if fin acknowledged
			c.UpdateState(CLOSED)
			continue
		case TIME_WAIT:
			// TODO handle remote fin
		}

		if segment.header.flags&TCP_URG != 0 {
			switch c.state {
			case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2:
				// TODO handle urg
			}
			continue
		}

		if segment.header.flags&TCP_FIN != 0 {
			switch c.state {
			case CLOSED, LISTEN, SYN_SENT:
				continue
			}

			// TODO notify user of the connection closing
			c.ackNum += segment.getPayloadSize()

			err := c.SendAck(c.seqNum, c.ackNum)
			logs.Info.Println("Sent ACK data in response to FIN")
			if err != nil {
				logs.Error.Println(err)
				continue
			}
			continue
		}

		switch c.state {
		case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2:
			c.recvBuffer = append(c.recvBuffer, segment.payload...)
			// TODO handle push flag
			// TODO adjust rcv.wnd, for now just multiplying by 2
			c.curWindow *= 2
			pay_size := segment.getPayloadSize()
			logs.Trace.Println("Payload Size is ", pay_size)
			c.ackNum += pay_size
			// TODO piggyback this

			err := c.SendAck(c.seqNum, c.ackNum)
			logs.Info.Println("Sent ACK data")
			if err != nil {
				logs.Error.Println(err)
				continue
			}
			continue
		case CLOSE_WAIT, CLOSING, LAST_ACK, TIME_WAIT:
			// should not occur, so drop packet
			continue
		}
	}
}

func (c *TCB) DealClosed(d *TCP_Packet) {
	if d.header.flags&TCP_RST != 0 {
		return
	}
	var seqNum uint32
	var ackNum uint32
	rstFlags := uint8(TCP_RST)
	if d.header.flags&TCP_ACK == 0 {
		seqNum = 0
		ackNum = d.header.seq + d.getPayloadSize()
		rstFlags = rstFlags | TCP_ACK
	} else {
		seqNum = d.header.ack
		ackNum = 0
	}

	rst_packet := &TCP_Packet{
		header: &TCP_Header{
			seq:     seqNum,
			ack:     ackNum,
			flags:   rstFlags,
			urg:     0,
			options: []byte{},
		},
		payload: []byte{},
	}

	err := c.SendPacket(rst_packet)
	logs.Info.Println("Sent ACK data")
	if err != nil {
		logs.Error.Println(err)
		return
	}
}

func (c *TCB) DealListen(d *TCP_Packet) {
	if d.header.flags&TCP_RST != 0 {
		return
	}
	if d.header.flags&TCP_ACK != 0 {
		err := c.SendReset(d.header.ack, 0)
		logs.Trace.Println("Sent ACK data")
		if err != nil {
			logs.Error.Println(err)
			return
		}
	}

	if d.header.flags&TCP_SYN != 0 {
		// TODO check security/compartment, if not match, send <SEQ=SEG.ACK><CTL=RST>
		// TODO handle SEG.PRC > TCB.PRC stuff
		// TODO if SEG.PRC < TCP.PRC continue
		c.ackNum = d.header.seq + 1
		c.IRS = d.header.seq
		// TODO queue other controls

		syn_ack_packet := &TCP_Packet{
			header: &TCP_Header{
				seq:     c.seqNum,
				ack:     c.ackNum,
				flags:   TCP_SYN | TCP_ACK,
				urg:     0,
				options: []byte{},
			},
			payload: []byte{},
		}

		err := c.SendPacket(syn_ack_packet)
		if err != nil {
			logs.Error.Println(err)
			return
		}
		logs.Trace.Println("Sent ACK data")

		c.seqNum += 1
		c.recentAckNum = c.ISS
		c.UpdateState(SYN_RCVD)
		return
	}
}

func (c *TCB) DealSynSent(d *TCP_Packet) {
	logs.Trace.Println("Dealing state syn-sent")
	if d.header.flags&TCP_ACK != 0 {
		logs.Trace.Println("verifing the ack")
		if d.header.flags&TCP_RST != 0 {
			return
		}
		if d.header.ack <= c.ISS || d.header.ack > c.seqNum {
			logs.Info.Println("Sending reset")
			err := c.SendReset(d.header.ack, 0)
			if err != nil {
				logs.Error.Println(err)
				return
			}
			return
		}
		if !(c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum) {
			logs.Error.Println("Incoming packet's ack is bad")
			return
		}

		// kill the retransmission
		err := c.UpdateLastAck(d.header.ack)
		if err != nil {
			logs.Error.Println(err)
			return
		}
	}

	if d.header.flags&TCP_RST != 0 {
		logs.Error.Println("error: connection reset")
		c.UpdateState(CLOSED)
		return
	}

	// TODO verify security/precedence

	if d.header.flags&TCP_SYN != 0 {
		logs.Trace.Println("rcvd a SYN")
		c.ackNum = d.header.seq + 1
		c.IRS = d.header.seq

		if d.header.flags&TCP_ACK != 0 {
			c.UpdateLastAck(d.header.ack)
			logs.Trace.Println("recentAckNum:", c.recentAckNum)
			logs.Trace.Println("ISS:", c.ISS)
		}

		if c.recentAckNum > c.ISS {
			logs.Trace.Println("rcvd a SYN-ACK")
			// the syn has been ACKed
			// reply with an ACK
			err := c.SendAck(c.seqNum, c.ackNum)
			if err != nil {
				logs.Error.Println(err)
			}

			c.UpdateState(ESTABLISHED)
			logs.Info.Println("Connection established")
			return
		} else {
			// special case... TODO deal with this case later
			// http://www.tcpipguide.com/free/t_TCPConnectionEstablishmentProcessTheThreeWayHandsh-4.htm
			// (Simultaneous Open Connection Establishment)

			//c.UpdateState(SYN_RCVD)
			// TODO send <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
			// TODO if more controls/txt, continue processing after established
		}
	}

	// Neither syn nor rst set
	logs.Info.Println("Dropping packet with seq: ", d.header.seq, "ack: ", d.header.ack)
}

func (c *TCB) DealSynRcvd(d *TCP_Packet) {
	if d.header.flags&TCP_SYN != 0 {
		// TODO send reset
	}
	if d.header.flags&TCP_ACK != 0 {
		// TODO Check segment acknowledgement is acceptable
		c.UpdateState(ESTABLISHED)
	}
}

func (c *TCB) Send(data []byte) error { // a non-blocking send call
	c.sendBuffer = append(c.sendBuffer, data...)
	return nil // TODO: read and send from the send buffer
}

func (c *TCB) Recv(num uint64) ([]byte, error) {
	amt := min(num, uint64(len(c.recvBuffer)))
	data := c.recvBuffer[0:amt]
	c.recvBuffer = c.recvBuffer[amt:]
	return data, nil // TODO: error handling
}

func (c *TCB) Close() error {
	return nil // TODO: free manager read buffer and send fin/fin+ack/etc. Also kill timers with a wait group
}

// TODO: support a status call

func (c *TCB) Abort() error {
	// TODO: kill all timers
	// TODO: kill all long term processes
	// TODO: send a reset
	// TODO: delete the TCB + assoc. data, enter closed state
	return nil
}
