package tcp

import (
	"github.com/hsheth2/logs"
)

func (c *TCB) packetDealer() {
	// read each tcp packet and deal with it
	/*logs*/ logs.Trace.Println(c.hash(), "Packet Dealer starting")
	for {
		///*logs*/logs.Trace.Println(c.hash(), "Waiting for packets")
		segment, open := <-c.read
		if !open || c.getState() == fsmClosed {
			/*logs*/ logs.Trace.Println(c.hash(), "Terminating packetdealer")
			return
		}
		/*logs*/ logs.Trace.Println(c.hash(), "packetDealer received a packet:", segment.header, " in state:", c.getState())
		c.packetDeal(segment)
	}
}

func (c *TCB) packetDeal(segment *packet) {
	defer tcpRecover()

	// If the state is CLOSED (i.e., TCB does not exist) then
	if c.getState() == fsmClosed {
		c.rcvClosed(segment)
		return
	}
	tcpAssert(c.getState() != fsmClosed, "state is closed")

	// Check if listen, or syn-sent state
	switch c.getState() {
	case fsmListen:
		c.rcvListen(segment)
		return
	case fsmSynSent:
		c.dealSynSent(segment)
		return
	}

	tcpAssert(c.getState() != fsmListen && c.getState() != fsmSynSent, "state is listen/synsent")
	// first check sequence number pg 69 txt
	if len(segment.payload) == 0 {
		if c.getWindow() == 0 {
			// TODO special allowances pg 69 rfc
			logs.Warn.Println(c.hash(), "Dropping unacceptable packet")
			//	if not rst
			if segment.header.flags&flagRst == 0 {
				c.sendAck(c.seqNum, c.ackNum)
			}
			return
		} else if !(c.ackNum <= segment.header.seq && segment.header.seq < c.ackNum+uint32(c.getWindow())) {
			logs.Warn.Println(c.hash(), "Dropping unacceptable packet")
			//	if not rst
			if segment.header.flags&flagRst == 0 {
				c.sendAck(c.seqNum, c.ackNum)
			}
			return
		}
	} else if c.getWindow() == 0 {
		logs.Warn.Println(c.hash(), "Dropping unacceptable packet")
		//	if not rst
		if segment.header.flags&flagRst == 0 {
			c.sendAck(c.seqNum, c.ackNum)
		}
		return
	} else {
		if !((c.ackNum <= segment.header.seq && segment.header.seq < c.ackNum+uint32(c.getWindow())) || (c.ackNum <= segment.header.seq+uint32(len(segment.payload))-1) && segment.header.seq+uint32(len(segment.payload))-1 < c.ackNum+uint32(c.getWindow())) {
			logs.Warn.Println(c.hash(), "Dropping unacceptable packet")
			//	if not rst
			if segment.header.flags&flagRst == 0 {
				c.sendAck(c.seqNum, c.ackNum)
			}
			return
		}
	}

	// second, check the RST bit
	if segment.header.flags&flagRst != 0 {
		// TODO finish: page 70
		switch c.getState() {
		case fsmSynRcvd:
			// TODO connection refused
			return
		case fsmEstablished, fsmFinWait1, fsmFinWait2, fsmCloseWait:
			// TODO not done
			return
		case fsmClosing, fsmLastAck, fsmTimeWait:
			c.updateState(fsmClosed)
			return
		}
	}

	// TODO third, check security/precedence
	// page 71

	// TODO check SYN (SYN bit shouldn't be there)
	if segment.header.flags&flagSyn != 0 {
		switch c.getState() {
		case fsmSynRcvd:
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
	if segment.header.flags&flagAck == 0 {
		logs.Warn.Println(c.hash(), "Dropping a packet without an ACK flag")
		return
	}
	tcpAssert(segment.header.flags&flagAck != 0, "segment missing ACK flag after verification")

	switch c.getState() {
	case fsmSynRcvd:
		c.dealSynRcvd(segment)
	case fsmEstablished:
		if c.recentAckNum < segment.header.ack && segment.header.ack <= c.seqNum {
			c.updateLastAck(segment.header.ack)
			// TODO handle retransmission queue
			// TODO update send window
		} else if c.recentAckNum > segment.header.ack {
			// ignore
			logs.Warn.Println(c.hash(), "Dropping packet: ACK validation failed")
			return
		} else if segment.header.ack > c.seqNum {
			// TODO send ack, drop segment, return
			logs.Warn.Println(c.hash(), "Dropping packet with bad ACK field")
			return
		}
	case fsmFinWait1:
		// TODO check if acknowledging FIN
		c.updateState(fsmFinWait2)
	case fsmFinWait2:
		//defer c.UpdateState(fsmClosed)
		// TODO if retransmission queue empty, acknowledge user's close with ok
	case fsmCloseWait:
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
	case fsmClosing:
		// TODO if ack is acknowledging our fin
		c.updateState(fsmTimeWait)
	// TODO else drop segment
	case fsmLastAck:
		// TODO if fin acknowledged
		c.updateState(fsmClosed)
		return
	case fsmTimeWait:
		// TODO handle remote fin
		c.updateState(fsmTimeWait)
		// This might be wrong
		err := c.sendAck(c.seqNum, c.ackNum)
		/*logs*/ logs.Trace.Println(c.hash(), "Sent ACK data in response to retrans FIN")
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}
	}

	if segment.header.flags&flagUrg != 0 {
		switch c.getState() {
		case fsmEstablished, fsmFinWait1, fsmFinWait2:
			// TODO handle urg
		}
		return
	}

	// TODO step 6, check URG bit

	// step 7 (?)
	switch c.getState() {
	case fsmEstablished, fsmFinWait1, fsmFinWait2:
		/*logs*/ logs.Trace.Println(c.hash(), "Received data of len:", len(segment.payload))
		c.recvBuffer = append(c.recvBuffer, segment.payload...)
		// TODO adjust rcv.wnd, for now just multiplying by 2
		if uint32(c.getWindow())*2 >= uint32(1)<<16 {
			c.windowMutex.Lock()
			c.curWindow *= 2
			c.windowMutex.Unlock()
		}
		payloadSize := segment.getPayloadSize()
		/*logs*/ logs.Trace.Println(c.hash(), "Payload Size is ", payloadSize)

		// TODO piggyback this

		if segment.header.flags&flagPsh != 0 {
			/*logs*/ logs.Trace.Println(c.hash(), "Pushing new data to client")
			c.pushData()
		}

		if payloadSize > 1 { // TODO make this correct
			c.seqAckMutex.Lock()
			c.ackNum += payloadSize
			c.seqAckMutex.Unlock()
			c.seqAckMutex.RLock()
			err := c.sendAck(c.seqNum, c.ackNum)
			c.seqAckMutex.RUnlock()
			/*logs*/ logs.Trace.Println(c.hash(), "Sent ACK data")
			if err != nil {
				logs.Error.Println(c.hash(), err)
				return
			}
		}
		//return
	case fsmCloseWait, fsmClosing, fsmLastAck, fsmTimeWait:
		// should not occur, so drop packet
		return
	}

	// eighth, check the FIN bit,
	if segment.header.flags&flagFin != 0 {
		switch c.getState() {
		case fsmClosed, fsmListen, fsmSynSent:
			// drop segment
			return
		}

		// TODO notify user of the connection closing
		c.seqAckMutex.Lock()
		c.ackNum += segment.getPayloadSize()
		c.seqAckMutex.Unlock()

		c.seqAckMutex.RLock()
		err := c.sendAck(c.seqNum, c.ackNum)
		c.seqAckMutex.RUnlock()
		/*logs*/ logs.Trace.Println(c.hash(), "Sent ACK data in response to FIN")
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}

		// FIN implies PSH
		/*logs*/ logs.Trace.Println(c.hash(), "Pushing data to client because of FIN")
		c.pushData()

		switch c.getState() {
		case fsmSynRcvd, fsmEstablished:
			c.updateState(fsmCloseWait)
		case fsmFinWait1:
			c.updateState(fsmTimeWait)
		case fsmFinWait2:
			c.updateState(fsmTimeWait)
		}
		return
	}
}

func (c *TCB) pushData() {
	// lock data mutex
	c.pushSignal.L.Lock()
	defer c.pushSignal.L.Unlock()

	// move data
	c.pushBuffer = append(c.pushBuffer, c.recvBuffer...)
	c.recvBuffer = []byte{}
	/*logs*/ logs.Trace.Println(c.hash(), "Pushing: new pushBuffer len:", len(c.pushBuffer))

	// signal push
	c.pushSignal.Signal()
}

func (c *TCB) rcvClosed(d *packet) {
	/*logs*/ logs.Trace.Println(c.hash(), "Dealing closed")
	if d.header.flags&flagRst != 0 {
		// drop incoming RSTs
		return
	}

	// respond with an RST
	var seqNum uint32
	var ackNum uint32
	var rstFlags flag = flagRst
	if d.header.flags&flagAck == 0 {
		seqNum = 0
		ackNum = d.header.seq + d.getPayloadSize()
		rstFlags = rstFlags | flagAck
	} else {
		seqNum = d.header.ack
		ackNum = 0
	}

	logs.Warn.Printf("%s Sending RST data with seq %d and ack %d because packet received in fsmClosed state", c.hash(), seqNum, ackNum)
	err := c.sendResetFlag(seqNum, ackNum, rstFlags)
	if err != nil {
		logs.Error.Println(c.hash(), err)
		return
	}
}

func (c *TCB) rcvListen(d *packet) {
	/*logs*/ logs.Trace.Println(c.hash(), "Dealing listen")

	if d.header.flags&flagRst != 0 {
		// drop incoming RSTs
		return
	}

	if d.header.flags&flagAck != 0 {
		err := c.sendReset(d.header.ack, 0)
		/*logs*/ logs.Trace.Println(c.hash(), "Sent ACK data")
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}
	}

	if d.header.flags&flagSyn != 0 {
		// TODO check security/compartment, if not match, send <SEQ=SEG.ACK><CTL=RST>
		// TODO handle SEG.PRC > TCB.PRC stuff
		// TODO if SEG.PRC < TCP.PRC continue
		c.ackNum = d.header.seq + 1
		c.irs = d.header.seq
		// TODO queue other controls

		// update state
		c.updateState(fsmSynRcvd)

		// send syn-ack response
		synAckPacket := &packet{
			header: &header{
				seq:     c.seqNum,
				ack:     c.ackNum,
				flags:   flagSyn | flagAck,
				urg:     0,
				options: []byte{},
			},
			payload: []byte{},
		}

		err := c.sendPacket(synAckPacket)
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}
		/*logs*/ logs.Trace.Println(c.hash(), "Sent ACK data")

		c.seqNum += 1
		c.recentAckNum = c.iss
		return
	}
}

func (c *TCB) dealSynSent(d *packet) {
	/*logs*/ logs.Trace.Println(c.hash(), "Dealing state syn-sent")
	if d.header.flags&flagAck != 0 {
		/*logs*/ logs.Trace.Println(c.hash(), "verifing the ack")
		if d.header.flags&flagRst != 0 {
			return
		}
		if d.header.ack <= c.iss || d.header.ack > c.seqNum {
			logs.Warn.Println(c.hash(), "Sending reset")
			err := c.sendReset(d.header.ack, 0)
			if err != nil {
				logs.Error.Println(c.hash(), err)
				return
			}
			return
		}
		if !(c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum) {
			logs.Error.Println(c.hash(), "Incoming packet's ack is bad")
			return
		}

		// kill the retransmission
		err := c.updateLastAck(d.header.ack)
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}
	}

	if d.header.flags&flagRst != 0 {
		logs.Error.Println(c.hash(), "error: connection reset")
		c.updateState(fsmClosed)
		return
	}

	// TODO verify security/precedence

	if d.header.flags&flagSyn != 0 {
		/*logs*/ logs.Trace.Println(c.hash(), "rcvd a SYN")
		c.ackNum = d.header.seq + 1
		c.irs = d.header.seq

		if d.header.flags&flagAck != 0 {
			c.updateLastAck(d.header.ack)
			/*logs*/ logs.Trace.Println(c.hash(), "recentAckNum:", c.recentAckNum)
			/*logs*/ logs.Trace.Println(c.hash(), "ISS:", c.iss)
		}

		if c.recentAckNum > c.iss {
			/*logs*/ logs.Trace.Println(c.hash(), "rcvd a SYN-ACK")
			// the syn has been ACKed
			// reply with an ACK
			c.updateState(fsmEstablished)
			c.seqAckMutex.RLock()
			err := c.sendAck(c.seqNum, c.ackNum)
			c.seqAckMutex.RUnlock()
			if err != nil {
				logs.Error.Println(c.hash(), err)
			}

			/*logs*/ logs.Trace.Println(c.hash(), "Connection established")
			return
		} else {
			// special case... TODO deal with this case later
			// http://www.tcpipguide.com/free/t_TCPConnectionEstablishmentProcessTheThreeWayHandsh-4.htm
			// (Simultaneous Open Connection Establishment)

			//c.UpdateState(fsmSynRcvd)
			// TODO send <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
			// TODO if more controls/txt, continue processing after established
		}
	}

	// Neither syn nor rst set
	logs.Warn.Println(c.hash(), "Dropping packet with seq: ", d.header.seq, "ack: ", d.header.ack)
}

func (c *TCB) dealSynRcvd(d *packet) {
	/*logs*/ logs.Trace.Println(c.hash(), "dealing Syn Rcvd")
	c.seqAckMutex.RLock()
	defer c.seqAckMutex.RUnlock()
	/*logs*/ logs.Trace.Printf("%s recentAck: %d, header ack: %d, seqNum: %d", c.hash(), c.recentAckNum, d.header.ack, c.seqNum)
	if c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum {
		/*logs*/ logs.Trace.Println(c.hash(), "SynRcvd -> Established")
		c.updateState(fsmEstablished)
	} else {
		err := c.sendReset(d.header.ack, 0)
		logs.Warn.Println(c.hash(), "Sent RST data")
		if err != nil {
			logs.Error.Println(c.hash(), err)
			return
		}
	}
}
