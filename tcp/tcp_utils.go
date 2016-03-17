package tcp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hsheth2/logs"
)

func (c *TCB) hash() string {
	return fmt.Sprintf("%d%d", c.lport, c.rport)
}

func (c *TCB) updateState(newState fsmState) {
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	c.updateStateReal(newState)
}

func (c *TCB) updateStateReal(newState fsmState) {
	/*logs*/logs.Trace.Println(c.hash(), "The New State is", newState)
	if c.state == fsmTimeWait && newState == fsmTimeWait {
		//		select {
		//		case c.timeWaitRestart <- true:
		//		default:
		//			/*logs*/logs.Trace.Println(c.hash(), "timeWaitRestart request already in progress; ignoring this request")
		//		}
		return
	} else if newState == fsmTimeWait && c.state != fsmClosed {
		// start timer
		go c.timeWaitTimer(c.timeWaitRestart)
	}
	c.state = newState
	c.stateUpdate.Broadcast()
	if c.serverParent != nil {
		go sendUpdate(c.serverParent.connQueueUpdate)
	}
}

func (c *TCB) getState() fsmState {
	c.stateUpdate.L.Lock()
	defer c.stateUpdate.L.Unlock()
	return c.state
}

func (c *TCB) getWindow() uint16 {
	c.windowMutex.RLock()
	defer c.windowMutex.RUnlock()
	return c.curWindow
}

func (c *TCB) updateLastAck(newAck uint32) error {
	/*logs*/logs.Trace.Println(c.hash(), "New ack number:", newAck)
	c.recentAckNum = newAck
	go c.recentAckUpdate.Broadcast(c.recentAckNum)
	return nil
}

func sendUpdate(update *sync.Cond) {
	update.L.Lock()
	update.Broadcast()
	update.L.Unlock()
}

func (c *TCB) timeWaitTimer(restart chan bool) error {
	select {
	case <-time.After(2 * time.Millisecond):
		close(c.timeWaitRestart)
		c.updateState(fsmClosed)
		return nil
	case <-restart:
		/*logs*/logs.Trace.Println(c.hash(), "Restarting timeWaitTimer")
		return c.timeWaitTimer(restart)
	}
}

func genRandSeqNum() (uint32, error) {
	x := make([]byte, 4) // four bytes
	_, err := rand.Read(x)
	if err != nil {
		return 0, errors.New("genRandSeqNum gave error:" + err.Error())
	}
	return uint32(x[0])<<24 | uint32(x[1])<<16 | uint32(x[2])<<8 | uint32(x[3]), nil
}

func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func tcpAssert(assert bool, msg string) {
	if !assert {
		panic("ASSERTION FAILED: " + msg)
	}
}

func tcpRecover() {
	if r := recover(); r != nil {
		logs.Error.Println("Recover from PANIC:", r)
	}
}
