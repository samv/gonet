package main

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
	//"time"
)

type TCB struct {
	read      chan []byte
	writer    *ipv4.RawConn
	ipAddress string // destination ip address
	srcIP     string // src ip address
	lport, rport uint16 // ports
	seqNum, ackNum    uint32 // sequence number
	state    uint // from the FSM
	kind     uint // type (server or client)
}

func New_TCB_From_Client(local, remote uint16, dstIP string) (*TCB, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	read, err := TCP_Port_Manager.bind(remote, local, dstIP)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", TCP_PROTO), dstIP) // only for read, not for write
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r, err := ipv4.NewRawConn(p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &TCB{
		lport: local,
		rport: remote,
		ipAddress: dstIP,
		srcIP: "127.0.0.1", // TODO: don't hardcode the srcIP
		read: read,
		writer: r,
		seqNum: genRandSeqNum(), // TODO verify that this works
		ackNum: uint32(0),    // Always 0 at start
		state: CLOSED,
		kind: TCP_CLIENT,
	}, nil
}

func (c *TCB) Connect() error { // TODO set the states for the FSM throughout
	// SYN
	window := uint16(43690) // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html

	SYN, _ := Make_TCP_Header(&TCP_Header{
		srcport: c.lport,
		dstport: c.rport,
		seq: c.seqNum,
		ack: c.ackNum,
		flags: TCP_SYN,
		window: window, // TODO improve the window size calculation
		urg: 0,
		options: []byte{0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0x0, 0x0, 0x0, 0x0, 0x01, 0x03, 0x03, 0x07,}, // TODO compute the options of SYN instead of hardcoding them
	}, c.ipAddress, c.srcIP)

	//c.writer.WriteTo(SYN)
	err := MyRawConnTCPWrite(c.writer, SYN, c.ipAddress)
	fmt.Println("Sent SYN")
	if err != nil {
		fmt.Println("Raw conn send", err)
	}

	// TODO: resend SYN on timeout

	// SYN-ACK // TODO set the state for the FSM
	// TODO also prepare for http://www.tcpipguide.com/free/t_TCPConnectionEstablishmentProcessTheThreeWayHandsh-4.htm (Simultaneous Open Connection Establishment)
	fmt.Println("Waiting for syn-ack")
	synack := <- c.read
	// TODO: verify the syn-ack flags (seq, ack, options-window scale, etc), and checksum
	fmt.Println("Syn-Ack Rcvd:", synack)

	// ACK
	c.seqNum++ // A+1
	B := uint32(synack[4]) << 24 | uint32(synack[5]) << 16 | uint32(synack[6]) << 8 | uint32(synack[7])
	c.ackNum = B+1

	ACK, _ := Make_TCP_Header(&TCP_Header{
		srcport: c.lport,
		dstport: c.rport,
		seq: c.seqNum,
		ack: c.ackNum,
		flags: TCP_ACK,
		window: window, // TODO improve the window field calculation
		urg: 0,
		options: []byte{},
	}, c.ipAddress, c.srcIP)

	err = MyRawConnTCPWrite(c.writer, ACK, c.ipAddress)
	fmt.Println("Sent ACK data")
	if err != nil {
		fmt.Println("Raw conn send", err)
	}
	// TODO set the state for the FSM

	return nil
}

func (c *TCB) Send(data []byte) error {
	return nil // TODO: implement TCP_Client send
}

func (c *TCB) Close() error {
	return nil // TODO: free manager read buffer
}
