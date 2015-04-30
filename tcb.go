package main

import (
	"errors"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
	"time"
)

type TCB struct {
	read            chan *TCP_Packet // input
	writer          *ipv4.RawConn    // output
	ipAddress       string           // destination ip address
	srcIP           string           // src ip address
	lport, rport    uint16           // ports
	seqNum          uint32           // seq number (SND.NXT)
	ackNum          uint32           // ack number (RCV.NXT)
	state           uint             // from the FSM
	stateUpdate     *sync.Cond       // signals when the state is changed
	kind            uint             // type (server or client)
	serverParent    *Server_TCB      // the parent server
	curWindow       uint16           // the current window size
	sendBuffer      []byte           // a buffer of bytes that need to be sent
	urgSendBuffer   []byte           // buffer of urgent data TODO urg data later
	recvBuffer      []byte           // bytes to pass to the application above
	resendDelay     time.Duration    // the delay before resending
	ISS             uint32           // the initial snd seq number
	IRS             uint32           // the initial rcv seq number
	recentAckNum    uint32           // the last ack received (also SND.UNA)
	recentAckUpdate *Notifier        // signals changes in recentAckNum
}

func New_TCB(local, remote uint16, dstIP string, read chan *TCP_Packet, write *ipv4.RawConn, kind uint) (*TCB, error) {
	Trace.Println("New_TCB")

	seq, err := genRandSeqNum()
	if err != nil {
		Error.Fatal(err)
		return nil, err
	}

	c := &TCB{
		lport:           local,
		rport:           remote,
		ipAddress:       dstIP,
		srcIP:           "127.0.0.1", // TODO: don't hardcode the srcIP
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
		recentAckUpdate: NewNotifier(),
	}
	Trace.Println("Starting the packet dealer")

	go c.PacketSender()
	go c.PacketDealer()

	return c, nil
}

func (c *TCB) UpdateState(newState uint) {
	c.state = newState
	go SendUpdate(c.stateUpdate)
	if c.serverParent != nil {
		go SendUpdate(c.serverParent.connQueueUpdate)
	}
}

func (c *TCB) UpdateLastAck(newAck uint32) error {
	Info.Println("Got an ack:", newAck)
	c.recentAckNum = newAck
	go SendNotifierBroadcast(c.recentAckUpdate, c.recentAckNum)
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
			Error.Println("Resend of packet seq", data.header.seq, "timed out")
			return errors.New("Resend timed out")
		}
	}
}

func (c *TCB) ListenForAck(successOut chan<- bool, end <-chan bool, targetAck uint32) {
	Trace.Println("Listening for ack:", targetAck)
	in := c.recentAckUpdate.Register(ACK_BUF_SZ)
	defer c.recentAckUpdate.Unregister(in)
	for {
		select {
		case v := <-in:
			if v.(uint32) == targetAck {
				return
			}
		case <-end:
			return
		}
	}
	successOut <- true
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
	d.header.window  = c.curWindow // TODO improve the window field calculation
	d.rip = c.ipAddress
	d.lip = c.srcIP

	pay, err := d.Marshal_TCP_Packet()
	if err != nil {
		Error.Println(err)
		return err
	}

	err = c.writer.WriteTo(&ipv4.Header{
		Version:  ipv4.Version,             // protocol version
		Len:      IP_HEADER_LEN,            // header length
		TOS:      0,                        // type-of-service (0 is everything normal)
		TotalLen: len(pay) + IP_HEADER_LEN, // packet total length (octets)
		ID:       0,                        // identification
		Flags:    ipv4.DontFragment,        // flags
		FragOff:  0,                        // fragment offset
		TTL:      DEFAULT_TTL,              // time-to-live (maximum lifespan in seconds)
		Protocol: TCP_PROTO,                // next protocol
		Checksum: 0,                        // checksum (autocomputed)
		Dst:      net.ParseIP(d.rip),       // destination address
	}, pay, nil)

	if err != nil {
		Error.Println(err)
		return err
	}

	return nil
}

func (c *TCB) SendReset(seq uint32, ack uint32) error {
	rst := &TCP_Packet{
		header: &TCP_Header{
			seq: seq,
			ack: ack,
			flags: TCP_RST,
			urg: 0,
			options: []byte{},
		},
		payload: []byte{},
	}

	return c.SendPacket(rst)
}

func (c *TCB) PacketDealer() {
	// read each tcp packet and deal with it
	Trace.Println("Packet Dealing")
	for {
		Trace.Println("Waiting for packets")
		segment := <-c.read
		Trace.Println("got a packet")

		// First check if closed, listen, or syn-sent state
		switch c.state {
		case CLOSED:
			Trace.Println("Dealing closed")
			c.DealClosed(segment)
			continue
		case LISTEN:
			Trace.Println("Dealing listen")
			c.DealListen(segment)
			continue
		case SYN_SENT:
			c.DealSynSent(segment)
			continue
		}

		// Otherwise

		// TODO check sequence number

		if segment.header.flags&TCP_RST != 0 {
			// TODO finish: page 70
			switch c.state {
			case SYN_RCVD:
				// not done
				continue
			case CLOSE_WAIT:
				// not done
				continue
			case TIME_WAIT:
				if segment.header.flags&TCP_RST != 0 {
					c.UpdateState(CLOSED)
				}
				continue
			}
		}

		// TODO check security/precedence
		// TODO check SYN (SYN bit shouldn't be there)

		if segment.header.flags&TCP_ACK == 0 {
			continue
		}

		switch c.state {
		case SYN_RCVD:
		case ESTABLISHED:

		}

		//		switch c.state {
		//		case CLOSED:
		//			Trace.Println("Dealing closed")
		//			go c.DealClosed(segment)
		//		case SYN_SENT:
		//			Trace.Println("Dealing syn-sent")
		//			go c.DealSynSent(segment)
		//		case SYN_RCVD:
		//			Trace.Println("Dealing syn-rcvd")
		//			go c.DealSynRcvd(segment)
		//		case ESTABLISHED:
		//			Trace.Println("Dealing established")
		//			go c.DealEstablished(segment)
		//		case FIN_WAIT_1:
		//			Trace.Println("Dealing Fin-Wait-1")
		//			go c.DealFinWaitOne(segment)
		//		case FIN_WAIT_2:
		//			go c.DealFinWaitTwo(segment)
		//		case CLOSE_WAIT:
		//			go c.DealCloseWait(segment)
		//		case CLOSING:
		//			go c.DealClosing(segment)
		//		case LAST_ACK:
		//			go c.DealLastAck(segment)
		//		case TIME_WAIT:
		//			go c.DealTimeWait(segment)
		//		default:
		//			Error.Println("Error: the current state is unknown")
		//		}
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
	Info.Println("Sent ACK data")
	if err != nil {
		Error.Println(err)
		return
	}
}

func (c *TCB) DealListen(d *TCP_Packet) {
	if d.header.flags&TCP_RST != 0 {
		return
	}
	if d.header.flags&TCP_ACK != 0 {
		/*RST, err := (&TCP_Header{
			srcport: c.lport,
			dstport: c.rport,
			seq:     d.header.ack,
			ack:     0,
			flags:   TCP_RST,
			window:  c.curWindow, // TODOold improve the window field calculation
			urg:     0,
			options: []byte{},
		}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
		if err != nil {
			Error.Println(err)
			return
		}

		err = MyRawConnTCPWrite(c.writer, RST, c.ipAddress)*/

		err := c.SendReset(d.header.ack, 0)
		Trace.Println("Sent ACK data")
		if err != nil {
			Error.Println(err)
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
			Error.Println(err)
			return
		}
		Trace.Println("Sent ACK data")

		c.seqNum += 1
		c.recentAckNum = c.ISS
		c.UpdateState(SYN_RCVD)
		return
	}
}

func (c *TCB) DealSynSent(d *TCP_Packet) {
	Trace.Println("Dealing state syn-sent")
	if d.header.flags&TCP_ACK != 0 {
		Trace.Println("verifing the ack")
		if d.header.flags&TCP_RST != 0 {
			return
		}
		if d.header.ack <= c.ISS || d.header.ack > c.seqNum {
			/*(RST, err := (&TCP_Header{
				srcport: c.lport,
				dstport: c.rport,
				seq:     d.header.ack,
				ack:     0,
				flags:   TCP_RST,
				window:  c.curWindow, // TODO improve the window field calculation
				urg:     0,
				options: []byte{},
			}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
			if err != nil {
				Error.Println(err)
				return
			}

			err = MyRawConnTCPWrite(c.writer, RST, c.ipAddress)*/

			err := c.SendReset(d.header.ack, 0)
			if err != nil {
				Error.Println(err)
				return
			}
			return
		}
		if !(c.recentAckNum <= d.header.ack && d.header.ack <= c.seqNum) {
			Error.Println("Incoming packet's ack is bad")
			return
		}
	}
	if d.header.flags&TCP_RST != 0 {
		Error.Println("error: connection reset")
		c.UpdateState(CLOSED)
		return
	}

	// TODO verify security/precedence

	if d.header.flags&TCP_SYN != 0 {
		Trace.Println("rcvd a SYN")
		c.ackNum = d.header.seq + 1
		c.IRS = d.header.seq

		if d.header.flags&TCP_ACK != 0 {
			c.UpdateLastAck(d.header.ack)
			Trace.Println("recentAckNum:", c.recentAckNum)
			Trace.Println("ISS:", c.ISS)
		}

		if c.recentAckNum > c.ISS {
			Trace.Println("rcvd a SYN-ACK")
			// the syn has been ACKed
			// reply with an ACK
			ack_packet := &TCP_Packet{
				header: &TCP_Header{
					seq:     c.seqNum,
					ack:     c.ackNum,
					flags:   TCP_ACK,
					urg:     0,
					options: []byte{},
				},
				payload: []byte{},
			}
			c.SendPacket(ack_packet)

			go c.UpdateState(ESTABLISHED)
			Info.Println("Connection established")
		} else {
			// special case... TODO deal with this case later
			// http://www.tcpipguide.com/free/t_TCPConnectionEstablishmentProcessTheThreeWayHandsh-4.htm (Simultaneous Open Connection Establishment)

			//c.UpdateState(SYN_RCVD)
			// TODO send <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
			// TODO if more controls/txt, continue processing after established
		}
	}

	// Neither syn nor rst set
	Trace.Println("Dropping packet")
	return
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

func (c *TCB) DealEstablished(d *TCP_Packet) {
	if d.header.flags&TCP_SYN != 0 {
		// TODO send reset
	}
	// TODO finish step 5 checks in rfc... I think we will need to split the packetDealer function into separate steps.

	c.UpdateLastAck(d.header.ack)

	// Send ACK
	c.seqNum++ // A+1
	B := d.header.seq
	c.ackNum = B + 1

	ack_packet := &TCP_Packet{
		header: &TCP_Header{
			seq:     c.seqNum,
			ack:     c.ackNum,
			flags:   TCP_ACK,
			urg:     0,
			options: []byte{},
		},
		payload: []byte{},
	}

	err := c.SendPacket(ack_packet)
	if err != nil {
		Error.Println(err)
		return
	}
	Info.Println("Sent ACK data")

	c.recvBuffer = append(c.recvBuffer, d.payload...)
}

func (c *TCB) DealFinWaitOne(d *TCP_Packet) {
	// TODO deal with Fin Wait 1
}

func (c *TCB) DealFinWaitTwo(d *TCP_Packet) {
	// TODO this function
}

func (c *TCB) DealCloseWait(d *TCP_Packet) {
	// TODO this function
}

func (c *TCB) DealClosing(d *TCP_Packet) {
	// TODO this function
}

func (c *TCB) DealLastAck(d *TCP_Packet) {
	// TODO this function
}

func (c *TCB) DealTimeWait(d *TCP_Packet) {
	// TODO this function
}

func (c *TCB) Send(data []byte) error { // a non-blocking send call
	c.sendBuffer = append(c.sendBuffer, data...)
	return nil // TODO: read and send from the send buffer
}

func (c *TCB) Recv(num uint64) ([]byte, error) {
	return nil, nil // TODO: implement TCB receive
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
