package tcp

import (
	"errors"
	"network/ipv4"
	"sync"
	"time"

	"network/ipv4/ipv4tps"

	"network/ipv4/ipv4src"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
)

type TCB struct {
	read             chan *TCP_Packet    // input
	writer           *ipv4.IP_Writer     // output
	ipAddress        *ipv4tps.IPaddress  // destination ip address
	srcIP            *ipv4tps.IPaddress  // src ip address
	lport, rport     uint16              // ports
	seqNum           uint32              // seq number (SND.NXT)
	ackNum           uint32              // ack number (RCV.NXT)
	seqAckMutex      *sync.RWMutex       // protects the seqNum and ackNum
	state            uint                // from the FSM
	timeWaitRestart  chan bool           // signals when the time_wait timer should restart
	stateUpdate      *sync.Cond          // signals when the state is changed
	kind             uint                // type (server or client)
	serverParent     *Server_TCB         // the parent server
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
	ISS              uint32              // the initial snd seq number
	IRS              uint32              // the initial rcv seq number
	recentAckNum     uint32              // the last ack received (also SND.UNA)
	recentAckUpdate  *notifiers.Notifier // signals changes in recentAckNum
	maxSegSize       uint16              // MSS (MTU)
}

func New_TCB(local, remote uint16, dstIP *ipv4tps.IPaddress, read chan *TCP_Packet, write *ipv4.IP_Writer, kind uint) (*TCB, error) {
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
		srcIP:            ipv4src.GlobalSource_IP_Table.Query(dstIP),
		read:             read,
		writer:           write,
		seqNum:           seq,
		ackNum:           uint32(0), // Always 0 at start
		seqAckMutex:      &sync.RWMutex{},
		state:            CLOSED,
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
		ISS:              seq,
		IRS:              0,
		recentAckNum:     0,
		recentAckUpdate:  notifiers.NewNotifier(),
		maxSegSize:       ipv4.MTU - TCP_BASIC_HEADER_SZ,
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

func (c *TCB) Recv(num uint64) ([]byte, error) { // blocking recv call
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
		//ch logs.Trace.Println(c.Hash(), "Waiting for push signal")
		c.pushSignal.Wait() // wait for a push
	}
	return nil, errors.New("Read failed")
}

const UINT32_MIN = uint32(0x00000000)
const UINT32_MAX = uint32(0xffffffff)

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
	if c.state == ESTABLISHED {
		//ch logs.Trace.Println(c.Hash(), "Entering fin-wait-1")
		c.updateStateReal(FIN_WAIT_1)
	} else if c.state == CLOSE_WAIT {
		//ch logs.Trace.Println(c.Hash(), "Entering last ack")
		c.updateStateReal(LAST_ACK)
	}
	c.stateUpdate.L.Unlock()

	// kill all retransmitter
	c.recentAckUpdate.Broadcast(UINT32_MIN)
	c.recentAckUpdate.Broadcast(c.ackNum)
	c.recentAckUpdate.Broadcast(UINT32_MAX)

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
		if c.state == CLOSED {
			break
		}
		c.stateUpdate.Wait()
	}
	//ch logs.Trace.Printf("%s Close of TCB with lport %d finished", c.Hash(), c.lport)

	//ch logs.Trace.Println(c.Hash(), "Unbinding TCB")
	err := TCP_Port_Manager.unbind(c.rport, c.lport, c.ipAddress)
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
