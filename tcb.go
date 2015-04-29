package main

import (
	"errors"
	"golang.org/x/net/ipv4"
	"sync"
	"time"
)

type TCB struct {
	read            chan *TCP_Packet
	writer          *ipv4.RawConn
	ipAddress       string        // destination ip address
	srcIP           string        // src ip address
	lport, rport    uint16        // ports
	seqNum, ackNum  uint32        // sequence number
	state           uint          // from the FSM
	stateUpdate     *sync.Cond    // signals when the state is changed
	kind            uint          // type (server or client)
	serverParent    *Server_TCB   // the parent server
	curWindow       uint16        // the current window size
	sendBuffer      []byte        // a buffer of bytes that need to be sent
	urgSendBuffer   []byte        // buffer of urgent data TODO urg data later
	recvBuffer      []byte        // bytes to pass to the application above
	resendDelay     time.Duration // the delay before resending
	recentAckNum    uint32        // the last ack received
	recentAckUpdate *Notifier     // signals changes in recentAckNum
}

func New_TCB(local, remote uint16, dstIP string, read chan *TCP_Packet, write *ipv4.RawConn, kind uint) (*TCB, error) {
	Trace.Println("New_TCB")
	c := &TCB{
		lport:           local,
		rport:           remote,
		ipAddress:       dstIP,
		srcIP:           "127.0.0.1", // TODO: don't hardcode the srcIP
		read:            read,
		writer:          write,
		seqNum:          genRandSeqNum(),
		ackNum:          uint32(0),       // Always 0 at start
		state:           CLOSED,
		stateUpdate:     sync.NewCond(&sync.Mutex{}),
		kind:            kind,
		serverParent:    nil,
		curWindow:       43690, // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
		resendDelay:     250 * time.Millisecond,
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

func SendNotifierBroadcast(update *Notifier, val interface{}) {
	update.Broadcast(val)
}

func (c *TCB) PacketSender() {
	// TODO: deal with data in send and urgSend buffers
}

func (c *TCB) SendWithRetransmit(data *TCP_Packet) error {
	// send the first packet
	packet, err := data.Marshal_TCP_Packet()
	if err != nil {
		return err
	}
	c.SendOnce(packet)

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
			c.SendOnce(packet)
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

func (c *TCB) SendOnce(pay []byte) error {
	return MyRawConnTCPWrite(c.writer, pay, c.ipAddress)
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
			return
		case LISTEN:
			Trace.Println("Dealing listen")
			c.DealListen(segment)
			return
		case SYN_SENT:
			c.DealSynSent(segment)
			return
		}

		// Otherwise
		// TODO left off on RFC pg 69

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

	RST, err := (&TCP_Header{
		srcport: c.lport,
		dstport: c.rport,
		seq:     seqNum,
		ack:     ackNum,
		flags:   rstFlags,
		window:  c.curWindow, // TODO improve the window field calculation
		urg:     0,
		options: []byte{},
	}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
	if err != nil {
		Error.Println(err)
		return
	}

	err = MyRawConnTCPWrite(c.writer, RST, c.ipAddress)
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
		RST, err := (&TCP_Header{
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

		err = MyRawConnTCPWrite(c.writer, RST, c.ipAddress)
		Trace.Println("Sent ACK data")
		if err != nil {
			Error.Println(err)
			return
		}
	}

	if d.header.flags&TCP_SYN != 0 {
		// TODO check security/comparment, if not match, send <SEQ=SEG.ACK><CTL=RST>
		// TODO handle SEG.PRC > TCB.PRC stuff
		// TODO if SEG.PRC < TCP.PRC continue
		// TODO Finish rest: pg 66 rfc
	}
}

func (c *TCB) DealSynSent(d *TCP_Packet) {
	Trace.Println("Dealing state syn-sent")
	ackAcceptable := true
	if d.header.flags&TCP_ACK != 0 {
		if d.header.flags&TCP_RST != 0 {
			return
		}
		// TODO if SEG.ACK =< ISS or SEG.ACK > SND.NXT send <SEQ=SEG.ACK><CTL=RST> and return
		// TODO if SND.UNA =< SEG.ACK =< SND.NXT then ackAcceptable = true, else false
	}
	if d.header.flags&TCP_RST != 0 {
		if ackAcceptable {
			Error.Println("error:connection reset")
			c.UpdateState(CLOSED)
		}
		return
	}

	// TODO check security/precedence

	if d.header.flags&TCP_SYN != 0 {
		// TODO set RCV.NXT to SEG.SEQ+1
		// TODO set IRS to SEG.SEQ
		// TODO set SND.UNA to SEG.ACK (if there is an ack)
		// TODO segments on retransmission queue should be removed

		// TODO if SND.UNA > ISS:
		// start if
		c.UpdateState(ESTABLISHED)
		// TODO send <SEQ=SND.NXT><ACK=RCV.NXT><CTL=ACK>
		// TODO if more controls/txt, continue processing at sixth step
		// end if, else:
		c.UpdateState(SYN_RCVD)
		// TODO send <SEQ=ISS><ACK=RCV.NXT><CTL=SYN,ACK>
		// TODO if more controls/txt, continue processing after established
		// end else
	} else {
		// Neither syn nor rst set
		return
	}

	/*if d.header.flags&TCP_SYN != 0 && d.header.flags&TCP_ACK != 0 {
		// received SYN-ACK
		Info.Println("Recieved syn-ack")

		// TODO: verify the seq and ack fields
		c.UpdateLastAck(d.header.ack)

		// Send ACK
		c.seqNum++ // A+1
		B := d.header.seq
		c.ackNum = B + 1
		ACK, err := (&TCP_Header{
			srcport: c.lport,
			dstport: c.rport,
			seq:     c.seqNum,
			ack:     c.ackNum,
			flags:   TCP_ACK,
			window:  c.curWindow, // TODO improve the window field calculation
			urg:     0,
			options: []byte{},
		}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
		if err != nil {
			Error.Println(err)
			return
		}

		err = MyRawConnTCPWrite(c.writer, ACK, c.ipAddress)
		Info.Println("Sent ACK data")
		if err != nil {
			Error.Println(err)
			return
		}
		c.UpdateState(ESTABLISHED)
	} else if d.header.flags&TCP_SYN == 0 {
		// TODO deal with special case: http://www.tcpipguide.com/free/t_TCPConnectionEstablishmentProcessTheThreeWayHandsh-4.htm (Simultaneous Open Connection Establishment)
	} else {
		// drop otherwise
	}*/
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

	ACK, err := (&TCP_Header{
		srcport: c.lport,
		dstport: c.rport,
		seq:     c.seqNum,
		ack:     c.ackNum,
		flags:   TCP_ACK,
		window:  c.curWindow, // TODO improve the window field calculation
		urg:     0,
		options: []byte{},
	}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
	if err != nil {
		Error.Println(err)
		return
	}

	err = MyRawConnTCPWrite(c.writer, ACK, c.ipAddress)
	Info.Println("Sent ACK data")
	if err != nil {
		Error.Println(err)
		return
	}

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
