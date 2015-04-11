package main

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
)

type TCP_Client_Manager struct {
	//reader *IP_Reader
	readBuffer map[uint16](map[string](chan []byte)) // dst protocol, ip
}

func New_TCP_Client_Manager() (*TCP_Client_Manager, error) {
	/*nr, err := NewNetwork_Reader() // TODO: create a global var for the network reader
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	cm := &TCP_Client_Manager{
		//reader: ipr,
		readBuffer: make(map[uint16](map[string](chan []byte))),
	}

	//go cm.readAll()

	return cm, nil
}

/*func (cm *TCP_Client_Manager) readAll() {
	for {
		ip, _, payload, err := cm.reader.ReadFrom()
		if err != nil {
			fmt.Println("TCP readAll error", err)
			continue
		}
		fmt.Println("Recved IP:", ip, "with payload", payload)
	}
}*/

type TCP_Client struct {
	manager   *TCP_Client_Manager
	//writer *IP_Writer
	conn   *ipv4.RawConn
	ipAddress string // destination ip address
	srcIP     string // src ip address
	src, dst  uint16 // ports
}

const (
	TCP_FIN = 0x01
	TCP_SYN = 0x02
	TCP_RSH = 0x04
	TCP_PSH = 0x08
	TCP_ACK = 0x10
	TCP_URG = 0x20
	TCP_ECE = 0x40
	TCP_CWR = 0x80
)

func (x *TCP_Client_Manager) New_TCP_Client(src, dst uint16, dstIP string) (*TCP_Client, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	p, err := net.ListenPacket("ip4:6", dstIP)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r, err := ipv4.NewRawConn(p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &TCP_Client{
		src: src,
		dst: dst,
		ipAddress: dstIP,
		srcIP: "127.0.0.1",
		manager: x,
		//writer: write,
		conn: r,
	}, nil
}

func (c *TCP_Client) Connect() error {
	// SYN
	initSeqNumber := uint32(3425) // Can be any number between 0 and 2^32 inclusive, TODO this needs to be changed later, it can't be predictable.
	initAckNumber := uint32(0)    // Always 0 at start
	window := uint16(43690)       // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
	headerLen := 40
	SYN := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte
		(byte)(c.dst >> 8), (byte)(c.dst), // Destination port in byte slice
		(byte)(initSeqNumber >> 24), (byte)(initSeqNumber >> 16), (byte)(initSeqNumber >> 8), (byte)(initSeqNumber),
		(byte)(initAckNumber >> 24), (byte)(initAckNumber >> 16), (byte)(initAckNumber >> 8), (byte)(initAckNumber),
		(byte)(
			(headerLen/4) << 4, // Size of header in 32 bit chunks. It is always 5 unless options are used. This is also the data offset.
			// bits 5-7 inclusive are reserved, always 0
			// bit 8 is flag 0(NS flag), set to 0 here because only SYN
		),
		(byte)(// Flags 1-8 inclusive
			// CWR
			// ECE
			// URG
			// ACK
			// PSH
			// RST
			0 | TCP_SYN, //SYN
			// FIN
		),
		(byte)(window >> 8), (byte)(window),
		0, 0, // TODO calc checksum, right now just set to 0
		0, 0, // URG pointer, only matters where URG flag is set.
		0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0, 0, 0, 0, 0x01, 0x03, 0x03, 0x07,
	}
	checksum := checksum(append(append(append(SYN, net.ParseIP(/*c.writer.src*/c.srcIP)...), net.ParseIP(/*c.writer.dst*/c.ipAddress)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(headerLen >> 8), byte(headerLen)}...))
	SYN[16] = byte(checksum >> 8)
	SYN[17] = byte(checksum)

	//c.writer.WriteTo(SYN)
	err := c.conn.WriteTo(&ipv4.Header{
		Version:  ipv4.Version, // protocol version
		Len:      20, // header length
		TOS:      0, // type-of-service (0 is everything normal)
		TotalLen: len(SYN) + 20, // packet total length (octets)
		ID:       0, // identification
		Flags:    ipv4.DontFragment, // flags
		FragOff:  0, // fragment offset
		TTL:      64, // time-to-live (maximum lifespan in seconds)
		Protocol: 6, // next protocol (17 is UDP)
		Checksum: 0, // checksum (apparently autocomputed)
		//Src:    net.IPv4(127, 0, 0, 1), // source address, apparently done automatically
		Dst: net.ParseIP(c.ipAddress), // destination address
		//Options                         // options, extension headers
	}, SYN, nil)
	fmt.Println("Sent data")
	if err != nil {
		fmt.Println("Raw conn send", err)
	}

	// TODO Wait for SYN + ACK, send back ACK

	return nil
}

func (c *TCP_Client) Send(data []byte) error {
	return nil // TODO: implement TCP_Client send
}

func (c *TCP_Client) Close() error {
	return nil // TODO: free manager read buffer
}
