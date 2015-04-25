package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
)

type Server_TCB struct {
	listener        chan *TCP_Packet
	listenPort      uint16
	listenIP        string
	state           uint
	kind            uint
	connQueue       chan *TCB
	connQueueUpdate *sync.Cond
}

func New_Server_TCB() (*Server_TCB, error) {
	x := &Server_TCB{
		listener:        nil,
		listenPort:      0,
		listenIP:        "*",
		state:           CLOSED,
		kind:            TCP_SERVER,
		connQueue:       make(chan *TCB, TCP_LISTEN_QUEUE_SZ),
		connQueueUpdate: sync.NewCond(&sync.Mutex{}),
	}

	return x, nil
}

func (s *Server_TCB) BindListen(port uint16, ip string) error {
	s.listenPort = port
	s.listenIP = ip
	read, err := TCP_Port_Manager.bind(port, 0, ip)
	if err != nil {
		return err
	}
	s.listener = read
	s.state = LISTEN

	go s.LongListener()

	return nil
}

func (s *Server_TCB) LongListener() {
	fmt.Println("Listener routine")
	for {
		in := <-s.listener
		if in.header.flags&TCP_RST != 0 {
			continue // parent TCB drops it
		} else if in.header.flags&TCP_ACK != 0 {
			// TODO send reset
		} else if in.header.flags&TCP_SYN == 0 {
			// TODO send reset
		}

		go func(s *Server_TCB, in *TCP_Packet) {
			lp := s.listenPort
			rp := in.header.srcport
			rIP := in.lip

			read, err := TCP_Port_Manager.bind(lp, rp, rIP)
			if err != nil {
				fmt.Println(err)
				return
			}

			p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", TCP_PROTO), rIP) // only for read, not for write
			if err != nil {
				fmt.Println(err)
				return
			}

			r, err := ipv4.NewRawConn(p)
			if err != nil {
				fmt.Println(err)
				return
			}

			c, err := New_TCB(lp, rp, rIP, read, r, TCP_SERVER)
			if err != nil {
				fmt.Println(err)
				return
			}
			c.serverParent = s

			// send syn-ack
			c.ackNum = in.header.ack + 1
			synack, err := (&TCP_Header{
				srcport: c.lport,
				dstport: c.rport,
				seq:     c.seqNum,
				ack:     c.ackNum,
				flags:   TCP_SYN | TCP_ACK,
				window:  c.curWindow,
				urg:     0,
				options: []byte{0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0x0, 0x0, 0x0, 0x0, 0x01, 0x03, 0x03, 0x07},
				// TODO compute the options of Syn-Ack instead of hardcoding them
			}).Marshal_TCP_Header(c.ipAddress, c.srcIP)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = MyRawConnTCPWrite(c.writer, synack, c.ipAddress)
			if err != nil {
				fmt.Println(err)
				return
			}

			c.UpdateState(SYN_RCVD)

			select {
			case s.connQueue <- c:
			default:
				// TODO send a reset
				fmt.Println(errors.New("ERR: listen queue is full"))
				return
			}
			return
		}(s, in)
	}
}

func (s *Server_TCB) Accept() (c *TCB, rip string, rport uint16, err error) {
	s.connQueueUpdate.L.Lock()
	defer s.connQueueUpdate.L.Unlock()
	for {
		// TODO add a timeout
		for i := 0; i < len(s.connQueue); i++ {
			next := <-s.connQueue
			if next.state == ESTABLISHED {
				return next, next.ipAddress, next.rport, nil
			}
			s.connQueue <- next
		}
		s.connQueueUpdate.Wait()
	}
}

func (s *Server_TCB) Close() error {
	return nil
}
