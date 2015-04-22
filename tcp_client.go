package main

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
	"errors"
	//"time"
)

type TCP_Main struct {
	tcp_reader *IP_Reader
	readBuffer map[uint16](map[uint16](map[string](chan []byte))) // dst, src port, ip
}

func New_TCP_Main() (*TCP_Main, error) {
	nr, err := NewNetwork_Reader() // TODO: create a global var for the network reader
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", TCP_PROTO)
	if err != nil {
		return nil, err
	}

	cm := &TCP_Main{
		tcp_reader: ipr,
		readBuffer: make(map[uint16](map[uint16](map[string](chan []byte)))),
	}

	go cm.readAll()

	return cm, nil
}

func (cm *TCP_Main) bind(srcport, dstport uint16, ip string) (chan []byte, error) {
	if _, ok := cm.readBuffer[dstport]; !ok {
		cm.readBuffer[dstport] = make(map[uint16](map[string](chan []byte)))
	}

	if _, ok := cm.readBuffer[dstport][srcport]; !ok {
		cm.readBuffer[dstport][srcport] = make(map[string](chan []byte))
	}

	if _, ok := cm.readBuffer[dstport][srcport][ip]; ok {
		return nil, errors.New("Ports and IP already binded to")
	}

	cm.readBuffer[dstport][srcport][ip] = make(chan []byte, TCP_INCOMING_BUFF_SZ)
	return cm.readBuffer[dstport][srcport][ip], nil
}

func (cm *TCP_Main) readAll() {
	for {
		ip, _, payload, err := cm.tcp_reader.ReadFrom()
		if err != nil {
			fmt.Println("TCP readAll error", err)
			continue
		}
		dstport := uint16(payload[0]) << 8 | uint16(payload[1]) // reversed to account
		srcport := uint16(payload[2]) << 8 | uint16(payload[3]) // for server sending

		if _, ok := cm.readBuffer[dstport]; ok {
			if _, ok := cm.readBuffer[dstport][srcport]; ok {
				if _, ok := cm.readBuffer[dstport][srcport][ip]; ok {
					go func() { cm.readBuffer[dstport][srcport][ip] <- payload }()
					continue
				} else if _, ok := cm.readBuffer[dstport][srcport]["*"]; ok {
					go func() { cm.readBuffer[dstport][srcport]["*"] <- payload }()
					continue
				}
			}
		}
		//fmt.Println(errors.New("Dst/Src port + ip not binded to"))
	}
}

type TCB struct {
	manager   *TCP_Main
	read      chan []byte
	writer    *ipv4.RawConn
	ipAddress string // destination ip address
	srcIP     string // src ip address
	src, dst  uint16 // ports
	seqNum, ackNum    uint32 // sequence number
	state    uint
}

func (x *TCP_Main) New_TCB(src, dst uint16, dstIP string) (*TCB, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	read, err := x.bind(src, dst, dstIP)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", TCP_PROTO), dstIP)
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
		src: src,
		dst: dst,
		ipAddress: dstIP,
		srcIP: "127.0.0.1", // TODO: don't hardcode the srcIP
		manager: x,
		read: read,
		writer: r,
		seqNum: uint32(3425), // Can be any number between 0 and 2^32 inclusive, TODO this needs to be changed later, it can't be predictable.
		ackNum: uint32(0),    // Always 0 at start
		state: CLOSED,
	}, nil
}

func (c *TCB) Connect() error {
	// SYN
	window := uint16(43690) // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html

	SYN, _ := Make_TCP_Header(&TCP_Header{
		srcport: c.src,
		dstport: c.dst,
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

	// SYN-ACK
	fmt.Println("Waiting for syn-ack")
	synack := <- c.read
	// TODO: verify the syn-ack flags (seq, ack, options-window scale, etc), and checksum
	fmt.Println("Syn-Ack Rcvd:", synack)

	// ACK
	c.seqNum++ // A+1
	B := uint32(synack[4]) << 24 | uint32(synack[5]) << 16 | uint32(synack[6]) << 8 | uint32(synack[7])
	c.ackNum = B+1

	ACK, _ := Make_TCP_Header(&TCP_Header{
		srcport: c.src,
		dstport: c.dst,
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

	return nil
}

func (c *TCB) Send(data []byte) error {
	return nil // TODO: implement TCP_Client send
}

func (c *TCB) Close() error {
	return nil // TODO: free manager read buffer
}
