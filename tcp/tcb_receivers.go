package tcp

import (
	"github.com/hsheth2/logs"
)

func (c *TCB) packetDealer() {
	// read each tcp packet and deal with it
	logs.Trace.Println(c.Hash(), "Packet Dealer starting")
	for {
		//logs.Trace.Println(c.Hash(), "Waiting for packets")
		segment := <-c.read
		logs.Trace.Println(c.Hash(), "packetDealer received a packet:", segment.header, " in state:", c.state)
		c.packetDeal(segment)
	}
}

func (c *TCB) packetDeal(segment *TCP_Packet) {
	defer Recover()

	// If the state is CLOSED (i.e., TCB does not exist) then
	if c.getState() == CLOSED {
		c.rcvClosed(segment)
		return
	}
	Assert(c.getState() != CLOSED, "state is closed")

	// Check if listen, or syn-sent state
	switch c.getState() {
	case LISTEN:
		c.rcvListen(segment)
		return
	case SYN_SENT:
		c.dealSynSent(segment)
		return
	}

	// first check sequence number pg 69 txt
	// TODO check sequence number

	// second, check the RST bit
	if segment.header.flags&TCP_RST != 0 {
		// TODO finish: page 70
		switch c.getState() {
		case SYN_RCVD:
			// TODO connection refused
			return
		case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2, CLOSE_WAIT:
			// TODO not done
			return
		case CLOSING, LAST_ACK, TIME_WAIT:
			c.UpdateState(CLOSED)
			return
		}
	}

	// TODO third, check security/precedence
	// page 71

	// TODO check SYN (SYN bit shouldn't be there)
	if segment.header.flags&TCP_SYN != 0 {
		switch c.getState() {
		case SYN_RCVD:
			if c.recentAckNum <= segment.header.ack && segment.header.ack <= c.seqNum {
				// in window, so it is an error
				// TODO send a reset
			} else {
				// TODO resend SYN-ACK
			}
		}
		return
	}

	// fifth, check the ACK field
	if segment.header.flags&TCP_ACK == 0 {
		logs.Info.Println(c.Hash(), "Dropping a packet without an ACK flag")
		return
	} else {
		Assert(segment.header.flags&TCP_ACK != 0, "segment missing ACK flag after verification")

		switch c.getState() {
		case SYN_RCVD:
			c.dealSynRcvd(segment)
		case ESTABLISHED:
			if c.recentAckNum < segment.header.ack && segment.header.ack <= c.seqNum {
				c.UpdateLastAck(segment.header.ack)
				// TODO handle retransmission queue
				// TODO update send window
			} else if c.recentAckNum > segment.header.ack {
				// ignore
				logs.Info.Println(c.Hash(), "Dropping packet: ACK validation failed")
				return
			} else if segment.header.ack > c.seqNum {
				// TODO send ack, drop segment, return
				logs.Info.Println(c.Hash(), "Dropping packet with bad ACK field")
				return
			}
		case FIN_WAIT_1:
			// TODO check if acknowledging FIN
			c.UpdateState(FIN_WAIT_2)
		case FIN_WAIT_2:
			//defer c.UpdateState(CLOSED)
			// TODO if retransmission queue empty, acknowledge user's close with ok
		case CLOSE_WAIT:
			if c.recentAckNum < segment.header.ack && segment.header.ack <= c.seqNum {
				c.recentAckNum = segment.header.ack
				// TODO handle retransmission queue
				// TODO update send window
			} else if c.recentAckNum > segment.header.ack {
				// ignore
				return
			} else if segment.header.ack > c.seqNum {
				// TODO send ack, drop segment, return
				return
			}
		case CLOSING:
			// TODO if ack is acknowledging our fin
			c.UpdateState(TIME_WAIT)
		// TODO else drop segment
		case LAST_ACK:
			// TODO if fin acknowledged
			c.UpdateState(CLOSED)
			return
		case TIME_WAIT:
			// TODO handle remote fin
			c.UpdateState(TIME_WAIT)
			// This might be wrong
			err := c.sendAck(c.seqNum, c.ackNum)
			logs.Info.Println(c.Hash(), "Sent ACK data in response to retrans FIN")
			if err != nil {
				logs.Error.Println(c.Hash(), err)
				return
			}
		}

		if segment.header.flags&TCP_URG != 0 {
			switch c.getState() {
			case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2:
				// TODO handle urg
			}
			return
		}

		// TODO step 6, check URG bit

		// step 7 (?)
		switch c.getState() {
		case ESTABLISHED, FIN_WAIT_1, FIN_WAIT_2:
			logs.Trace.Println(c.Hash(), "Received data of len:", len(segment.payload))
			c.recvBuffer = append(c.recvBuffer, segment.payload...)
			// TODO adjust rcv.wnd, for now just multiplying by 2
			if uint32(c.curWindow)*2 >= uint32(1)<<16 {
				c.curWindow *= 2
			}
			pay_size := segment.getPayloadSize()
			logs.Trace.Println(c.Hash(), "Payload Size is ", pay_size)

			// TODO piggyback this

			if segment.header.flags&TCP_PSH != 0 {
				logs.Trace.Println(c.Hash(), "Pushing new data to client")
				c.pushData()
			}

			if pay_size > 1 { // TODO make this correct
				c.seqAckMutex.Lock()
				c.ackNum += pay_size
				c.seqAckMutex.Unlock()
				c.seqAckMutex.RLock()
				err := c.sendAck(c.seqNum, c.ackNum)
				c.seqAckMutex.RUnlock()
				logs.Info.Println(c.Hash(), "Sent ACK data")
				if err != nil {
					logs.Error.Println(c.Hash(), err)
					return
				}
			}
			//return
		case CLOSE_WAIT, CLOSING, LAST_ACK, TIME_WAIT:
			// should not occur, so drop packet
			return
		}

		// eighth, check the FIN bit,
		if segment.header.flags&TCP_FIN != 0 {
			switch c.getState() {
			case CLOSED, LISTEN, SYN_SENT:
				// drop segment
				return
			}

			// TODO notify user of the connection closing
			c.ackNum += segment.getPayloadSize()

			err := c.sendAck(c.seqNum, c.ackNum)
			logs.Info.Println(c.Hash(), "Sent ACK data in response to FIN")
			if err != nil {
				logs.Error.Println(c.Hash(), err)
				return
			}

			// FIN implies PSH
			logs.Trace.Println(c.Hash(), "Pushing data to client because of FIN")
			c.pushData()

			switch c.getState() {
			case SYN_RCVD, ESTABLISHED:
				c.UpdateState(CLOSE_WAIT)
			case FIN_WAIT_1:
				c.UpdateState(CLOSED)
			case FIN_WAIT_2:
				c.UpdateState(CLOSED)
			}
			return
		}
	}
}

func (c *TCB) pushData() {
	// lock data mutex
	c.pushSignal.L.Lock()
	defer c.pushSignal.L.Unlock()

	// move data
	c.pushBuffer = append(c.pushBuffer, c.recvBuffer...)
	c.recvBuffer = []byte{}
	logs.Trace.Println(c.Hash(), "Pushing: new pushBuffer len:", len(c.pushBuffer))

	// signal push
	c.pushSignal.Signal()
}

func (c *TCB) rcvClosed(d *TCP_Packet) {
	logs.Trace.Println(c.Hash(), "Dealing closed")
	if d.header.flags&TCP_RST != 0 {
		// drop incoming RSTs
		return
	}

	// respond with an RST
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

	logs.Info.Printf("%s Sending RST data with seq %d and ack %d because packet received in CLOSED state", c.Hash(), seqNum, ackNum)
	err := c.sendResetFlag(seqNum, ackNum, rstFlags)
	if err != nil {
		logs.Error.Println(c.Hash(), err)
		return
	}
}

func (c *TCB) rcvListen(d *TCP_Packet) {
	logs.Trace.Println(c.Hash(), "Dealing listen")

	if d.header.flags&TCP_RST != 0 {
		// drop incoming RSTs
		return
	}

	if d.header.flags&TCP_ACK != 0 {
		err := c.sendReset(d.header.ack, 0)
		logs.Trace.Println(c.Hash(), "Sent ACK data")
		if err != nil {
			logs.Error.Println(c.Hash(), err)
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

		// update state
		c.UpdateState(SYN_RCVD)

		// send syn-ack response
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

		err := c.sendPacket(syn_ack_packet)
		if err != nil {
			logs.Error.Println(c.Hash(), err)
			return
		}
		logs.Trace.Println(c.Hash(), "Sent ACK data")

		c.seqNum += 1
		c.recentAckNum = c.ISS
		return
	}
}

func (c *TCB) dealSynSent(d *TCP_Packet) {
	logs.Trace.Println(c.Hash(), "Dealing state syn-sent")
	if d.header.flags&TCP_ACK != 0 {
		logs.Trace.Println(c.Hash(), "verifing the ack")
		if d.header.flags&TCP_RST != 0 {
			return
		}
		if d.header.ack <= c.ISS || d.header.ack > c.seqNum {
			logs.Info.Println(c.Hash(), "Sending reset")
			err := c.sendReset(d.header.ack, 0)
			if err != nil {
				logs.Error.Println(c.Hash(), err)
				return
			}
			return
		}
		if !(c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum) {
			logs.Error.Println(c.Hash(), "Incoming packet's ack is bad")
			return
		}

		// kill the retransmission
		err := c.UpdateLastAck(d.header.ack)
		if err != nil {
			logs.Error.Println(c.Hash(), err)
			return
		}
	}

	if d.header.flags&TCP_RST != 0 {
		logs.Error.Println(c.Hash(), "error: connection reset")
		c.UpdateState(CLOSED)
		return
	}

	// TODO verify security/precedence

	if d.header.flags&TCP_SYN != 0 {
		logs.Trace.Println(c.Hash(), "rcvd a SYN")
		c.ackNum = d.header.seq + 1
		c.IRS = d.header.seq

		if d.header.flags&TCP_ACK != 0 {
			c.UpdateLastAck(d.header.ack)
			logs.Trace.Println(c.Hash(), "recentAckNum:", c.recentAckNum)
			logs.Trace.Println(c.Hash(), "ISS:", c.ISS)
		}

		if c.recentAckNum > c.ISS {
			logs.Trace.Println(c.Hash(), "rcvd a SYN-ACK")
			// the syn has been ACKed
			// reply with an ACK
			c.UpdateState(ESTABLISHED)
			err := c.sendAck(c.seqNum, c.ackNum)
			if err != nil {
				logs.Error.Println(c.Hash(), err)
			}

			logs.Info.Println(c.Hash(), "Connection established")
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
	logs.Info.Println(c.Hash(), "Dropping packet with seq: ", d.header.seq, "ack: ", d.header.ack)
}

func (c *TCB) dealSynRcvd(d *TCP_Packet) {
	logs.Trace.Println(c.Hash(), "dealing Syn Rcvd")
	c.seqAckMutex.RLock()
	defer c.seqAckMutex.RUnlock()
	logs.Trace.Printf("%s recentAck: %d, header ack: %d, seqNum: %d", c.Hash(), c.recentAckNum, d.header.ack, c.seqNum)
	if c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum {
		logs.Trace.Println(c.Hash(), "SynRcvd -> Established")
		c.UpdateState(ESTABLISHED)
	} else {
		err := c.sendReset(d.header.ack, 0)
		logs.Info.Println(c.Hash(), "Sent RST data")
		if err != nil {
			logs.Error.Println(c.Hash(), err)
			return
		}
	}
}
