package main

import (
	"net"
	"fmt"
)

type TCP_Client_Manager struct {
	reader *IP_Reader
	readBuffer map[uint16](map[string](chan []byte)) // dst protocol, ip
}

func New_TCP_Client_Manager() (*TCP_Client_Manager, error) {
	nr, err := NewNetwork_Reader() // TODO: create a global var for the network reader
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", TCP_PROTO)
	if err != nil {
		return nil, err
	}

	cm := &TCP_Client_Manager{
		reader: ipr,
		readBuffer: make(map[uint16](map[string](chan []byte))),
	}

	go cm.readAll()

	return cm, nil
}

func (cm *TCP_Client_Manager) readAll() {
	for {
		ip, _, payload, err := cm.reader.ReadFrom()
		if err != nil {
			fmt.Println("TCP readAll error", err)
			continue
		}
		fmt.Println("Recved IP:", ip, "with payload", payload)
	}
}

type TCP_Client struct {
	manager   *TCP_Client_Manager
	writer *IP_Writer
	ipAddress string // destination ip address
	src, dst  uint16 // ports
}

func (x *TCP_Client_Manager) New_TCP_Client(src, dst uint16, dstIP string) (*TCP_Client, error) {
	write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}
	return &TCP_Client{src: src, dst: dst, ipAddress: dstIP, manager: x, writer: write}, nil
}

func (c *TCP_Client) Connect() error {
	// SYN
	initSeqNumber := uint32(0) // Can be any number between 0 and 2^32 inclusive, TODO this needs to be changed later, it can't be predictable.
	initAckNumber := uint32(0) // Always 0 at start
	window := uint16(2 ^ 10)   // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
	SYN := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte
		(byte)(c.dst >> 8), (byte)(c.dst), // Destination port in byte slice
		(byte)(initSeqNumber >> 24), (byte)(initSeqNumber >> 16), (byte)(initSeqNumber >> 8), (byte)(initSeqNumber),
		(byte)(initAckNumber >> 24), (byte)(initAckNumber >> 16), (byte)(initAckNumber >> 8), (byte)(initAckNumber),
		(byte)(
			5 << 4, // Size of header in 32 bit chunks. It is always 5 unless options are used. This is also the data offset.
			// bits 5-7 inclusive are reserved, always 0
			// bit 8 is flag 0(NS flag), set to 0 here because only SYN
		),
		(byte)( // Flags 1-8 inclusive
			// CWR
			// ECE
			// URG
			// ACK
			// PSH
			// RST
			1 << 1, //SYN
			// FIN
		),
		(byte)(window >> 8), (byte)(window),
		0, 0, // TODO calc checksum, right now just set to 0
		0, 0, // URG pointer, only matters where URG flag is set.
	}
	checksum := checksum(append(append(append(SYN, net.ParseIP(c.writer.src)...), net.ParseIP(c.writer.dst)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(20 >> 8), byte(20)}...))
	SYN[16] = byte(checksum >> 8)
	SYN[17] = byte(checksum)

	c.writer.WriteTo(SYN)

	// TODO Wait for SYN + ACK, send back ACK

	return nil
}

func (c *TCP_Client) Send(data []byte) error {
	return nil // TODO: implement TCP_Client send
}

func (c *TCP_Client) Close() error {
	return nil // TODO: free manager read buffer
}
