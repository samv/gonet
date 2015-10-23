package tcp

import (
	"errors"
	"math"
	"network/ipv4"
	"sync"
	"time"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
)

type TCB struct {
	read             chan *packet        // input
	writer           ipv4.Writer         // output
	ipAddress        *ipv4.Address       // destination ip address
	srcIP            *ipv4.Address       // src ip address
	lport, rport     uint16              // ports
	seqNum           uint32              // seq number (SND.NXT)
	ackNum           uint32              // ack number (RCV.NXT)
	seqAckMutex      *sync.RWMutex       // protects the seqNum and ackNum
	state            fsmState            // from the FSM
	timeWaitRestart  chan bool           // signals when the time_wait timer should restart
	stateUpdate      *sync.Cond          // signals when the state is changed
	kind             tcbParentType       // type (server or client)
	serverParent     *Server             // the parent server
	curWindow        uint16              // the current window size
	windowMutex      *sync.RWMutex       // protects access to the window size
	sendBuffer       []byte              // a buffer of bytes that need to be sent
	urgSendBuffer    []byte              // buffer of urgent data TODO urg data later
	sendBufferUpdate *sync.Cond          // notifies of send buffer updates
	stopSending      bool                // if the send function is allowed
	sendFinished     *notifiers.Notifier // broadcast when done sending
	recvBuffer       []byte              // bytes to received but not yet pushed
	pushBuffer       []byte              // bytes to push to client
	pushSignal       *sync.Cond          // signals upon push
	resendDelay      time.Duration       // the delay before resending
	iss              uint32              // the initial snd seq number
	irs              uint32              // the initial rcv seq number
	recentAckNum     uint32              // the last ack received (also SND.UNA)
	recentAckUpdate  *notifiers.Notifier // signals changes in recentAckNum
	maxSegSize       uint16              // MSS (MTU)
}

func newTCB(local, remote uint16, dstIP *ipv4.Address, read chan *packet, write ipv4.Writer, kind tcbParentType) (*TCB, error) {
	//ch logs.Trace.Println("New_TCB")

	seq, err := genRandSeqNum()
	if err != nil {
		logs.Error.Fatal(err)
		return nil, err
	}

	c := &TCB{
		lport:            local,
		rport:            remote,
		ipAddress:        dstIP,
		srcIP:            ipv4.GlobalRoutingTable.Query(dstIP),
		read:             read,
		writer:           write,
		seqNum:           seq,
		ackNum:           uint32(0), // Always 0 at start
		seqAckMutex:      &sync.RWMutex{},
		state:            fsmClosed,
		stateUpdate:      sync.NewCond(&sync.Mutex{}),
		timeWaitRestart:  make(chan bool, 1),
		kind:             kind,
		serverParent:     nil,
		curWindow:        43690, // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
		windowMutex:      &sync.RWMutex{},
		sendBufferUpdate: sync.NewCond(&sync.Mutex{}),
		stopSending:      false,
		sendFinished:     notifiers.NewNotifier(),
		pushSignal:       sync.NewCond(&sync.Mutex{}),
		resendDelay:      250 * time.Millisecond,
		iss:              seq,
		irs:              0,
		recentAckNum:     0,
		recentAckUpdate:  notifiers.NewNotifier(),
		maxSegSize:       ipv4.IPMTU - basicHeaderSize,
	}
	////ch logs.Trace.Println("Starting the packet dealer")

	go c.packetSender()
	go c.packetDealer()

	return c, nil
}

func (c *TCB) Send(data []byte) error { // a blocking send call
	c.sendBufferUpdate.L.Lock()
	defer c.sendBufferUpdate.L.Unlock()
	if c.stopSending {
		return errors.New("Sending is not allowed anymore")
	}

	c.sendBuffer = append(c.sendBuffer, data...)
	c.sendBufferUpdate.Broadcast()
	return nil
}

func (c *TCB) Recv(num uint64) ([]byte, error) { // blocking recv call TODO add timeout
	c.pushSignal.L.Lock()
	defer c.pushSignal.L.Unlock()
	for {
		//ch logs.Trace.Println(c.Hash(), "Attempting to read off of pushBuffer")
		//ch logs.Trace.Println(c.Hash(), "Amt of data on pushBuffer:", len(c.pushBuffer))
		amt := min(num, uint64(len(c.pushBuffer)))
		if amt != 0 {
			data := c.pushBuffer[:amt]
			c.pushBuffer = c.pushBuffer[amt:]
			return data, nil
		}
		switch c.getState() {
		case fsmClosed, fsmLastAck, fsmCloseWait:
			return nil, errors.New("connection closed by remote; cannot receive")
		}
		//ch logs.Trace.Println(c.Hash(), "Waiting for push signal")
		c.pushSignal.Wait() // wait for a push
	}
	//return nil, errors.New("Read failed")
}

func (c *TCB) Close() error {
	//ch logs.Trace.Println(c.Hash(), "Closing TCB with lport:", c.lport)

	// block all future sends
	c.sendBufferUpdate.L.Lock()
	c.stopSending = true
	c.sendBufferUpdate.L.Unlock()

	if len(c.sendBuffer) != 0 {
		//ch logs.Trace.Println(c.Hash(), "Blocking until all pending writes complete")
		c.sendFinished.Wait() // wait for send to finish
	}

	// update state for sending FIN packet
	c.stateUpdate.L.Lock()
	if c.state == fsmEstablished {
		//ch logs.Trace.Println(c.Hash(), "Entering fin-wait-1")
		c.updateStateReal(fsmFinWait1)
	} else if c.state == fsmCloseWait {
		//ch logs.Trace.Println(c.Hash(), "Entering last ack")
		c.updateStateReal(fsmLastAck)
	}
	c.stateUpdate.L.Unlock()

	// kill all retransmitters
	c.recentAckUpdate.Broadcast(uint32(0))
	c.recentAckUpdate.Broadcast(c.ackNum)
	c.recentAckUpdate.Broadcast(uint32(math.MaxUint32))

	// send FIN
	//ch logs.Trace.Println(c.Hash(), "Sending FIN within close")
	c.seqAckMutex.RLock()
	c.sendFin(c.seqNum, c.ackNum)
	c.seqAckMutex.RUnlock()
	c.seqAckMutex.Lock()
	c.seqNum += 1 // TODO make this not dumb
	c.seqAckMutex.Unlock()

	// wait until state becomes CLOSED
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	for {
		if c.state == fsmClosed {
			break
		}
		c.stateUpdate.Wait()
	}
	//ch logs.Trace.Printf("%s Close of TCB with lport %d finished", c.Hash(), c.lport)

	//ch logs.Trace.Println(c.Hash(), "Unbinding TCB")
	err := portManager.unbind(c.rport, c.lport, c.ipAddress)
	if err != nil {
		return err
	}

	c = nil
	return nil // TODO: free manager read buffer. Also kill timers with a wait group
}

// TODO: support a status call

func (c *TCB) Abort() error {
	// TODO: kill all timers
	// TODO: kill all long term processes
	// TODO: send a reset
	// TODO: delete the TCB + assoc. data, enter closed state
	return nil
}
