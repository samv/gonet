package tcpp

import (
	"errors"
	"network/ipv4p"
	"sync"
	"time"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
	"golang.org/x/net/ipv4"
)

type TCB struct {
	read             chan *TCP_Packet    // input
	writer           *ipv4.RawConn       // output
	ipAddress        string              // destination ip address
	srcIP            string              // src ip address
	lport, rport     uint16              // ports
	seqNum           uint32              // seq number (SND.NXT)
	ackNum           uint32              // ack number (RCV.NXT)
	state            uint                // from the FSM
	stateUpdate      *sync.Cond          // signals when the state is changed
	kind             uint                // type (server or client)
	serverParent     *Server_TCB         // the parent server
	curWindow        uint16              // the current window size
	sendBuffer       []byte              // a buffer of bytes that need to be sent
	urgSendBuffer    []byte              // buffer of urgent data TODO urg data later
	sendBufferUpdate *sync.Cond          // notifies of send buffer updates
	stopSending      bool                // if the send function is allowed
	sendFinished     *notifiers.Notifier // broadcast when done sending
	recvBuffer       []byte              // bytes to pass to the application above
	resendDelay      time.Duration       // the delay before resending
	ISS              uint32              // the initial snd seq number
	IRS              uint32              // the initial rcv seq number
	recentAckNum     uint32              // the last ack received (also SND.UNA)
	recentAckUpdate  *notifiers.Notifier // signals changes in recentAckNum
	maxSegSize       uint16              // MSS (MTU)
}

func New_TCB(local, remote uint16, dstIP string, read chan *TCP_Packet, write *ipv4.RawConn, kind uint) (*TCB, error) {
	logs.Trace.Println("New_TCB")

	seq, err := genRandSeqNum()
	if err != nil {
		logs.Error.Fatal(err)
		return nil, err
	}

	c := &TCB{
		lport:            local,
		rport:            remote,
		ipAddress:        dstIP,
		srcIP:            ipv4p.GetSrcIP(dstIP),
		read:             read,
		writer:           write,
		seqNum:           seq,
		ackNum:           uint32(0), // Always 0 at start
		state:            CLOSED,
		stateUpdate:      sync.NewCond(&sync.Mutex{}),
		kind:             kind,
		serverParent:     nil,
		curWindow:        43690, // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
		sendBufferUpdate: sync.NewCond(&sync.Mutex{}),
		stopSending:      false,
		sendFinished:     notifiers.NewNotifier(),
		resendDelay:      250 * time.Millisecond,
		ISS:              seq,
		IRS:              0,
		recentAckNum:     0,
		recentAckUpdate:  notifiers.NewNotifier(),
		maxSegSize:       ipv4p.MTU - TCP_BASIC_HEADER_SZ,
	}
	//logs.Trace.Println("Starting the packet dealer")

	go c.packetSender()
	go c.packetDealer()

	return c, nil
}

func (c *TCB) Send(data []byte) error { // a non-blocking send call
	if c.stopSending {
		return errors.New("Sending is not allowed anymore")
	}

	c.sendBuffer = append(c.sendBuffer, data...)
	go SendUpdate(c.sendBufferUpdate)
	return nil // TODO: read and send from the send buffer
}

func (c *TCB) Recv(num uint64) ([]byte, error) {
	amt := min(num, uint64(len(c.recvBuffer)))
	data := c.recvBuffer[0:amt]
	c.recvBuffer = c.recvBuffer[amt:]
	return data, nil // TODO: error handling
}

func (c *TCB) Close() error {
	logs.Trace.Println("Closing TCB with lport:", c.lport)
	c.stopSending = true // block all future sends
	if len(c.sendBuffer) != 0 {
		<-c.sendFinished.Register(1) // wait for send to finish
	}

	// send FIN
	logs.Info.Println("Sending FIN within close")
	c.sendFin(c.seqNum, c.ackNum)
	c.seqNum += 1 // TODO make this not dumb

	// update state accordingly
	if c.state == ESTABLISHED {
		c.UpdateState(FIN_WAIT_1)
	} else if c.state == CLOSE_WAIT {
		c.UpdateState(CLOSED)
	}

	// wait until state becomes CLOSED
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	for {
		if c.state == CLOSED {
			break
		}
		c.stateUpdate.Wait()
	}
	logs.Trace.Printf("Close of TCB with lport %d finished", c.lport)

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
