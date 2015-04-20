package main

import (
	"net"
	"fmt"
	"golang.org/x/net/ipv4"
	"errors"
	//"time"
)

const tcp_buff_sz = 10

type TCP_Client_Manager struct {
	tcp_reader *IP_Reader
	readBuffer map[uint16](map[uint16](map[string](chan []byte))) // dst, src port, ip
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
		tcp_reader: ipr,
		readBuffer: make(map[uint16](map[uint16](map[string](chan []byte)))),
	}

	go cm.readAll()

	return cm, nil
}

func (cm *TCP_Client_Manager) bind(srcport, dstport uint16, ip string) (chan []byte, error) {
	if _, ok := cm.readBuffer[dstport]; !ok {
		cm.readBuffer[dstport] = make(map[uint16](map[string](chan []byte)))
	}

	if _, ok := cm.readBuffer[dstport][srcport]; !ok {
		cm.readBuffer[dstport][srcport] = make(map[string](chan []byte))
	}

	if _, ok := cm.readBuffer[dstport][srcport][ip]; ok {
		return nil, errors.New("Ports and IP already binded to")
	}

	cm.readBuffer[dstport][srcport][ip] = make(chan []byte, tcp_buff_sz)
	return cm.readBuffer[dstport][srcport][ip], nil
}

func (cm *TCP_Client_Manager) readAll() {
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
					cm.readBuffer[dstport][srcport][ip] <- payload
					continue
				} else if _, ok := cm.readBuffer[dstport][srcport]["*"]; ok {
					cm.readBuffer[dstport][srcport]["*"] <- payload
					continue
				}
			}
		}
		//fmt.Println(errors.New("Dst/Src port + ip not binded to"))
	}
}

type TCP_Client struct {
	manager   *TCP_Client_Manager
	read      chan []byte
	writer    *ipv4.RawConn
	ipAddress string // destination ip address
	srcIP     string // src ip address
	src, dst  uint16 // ports
	seqNum, ackNum    uint32 // sequence number
	status    uint
}

const (
	LISTEN = 1
	SYN_SENT
	SYN_RCVD
	ESTABLISHED
	FIN_WAIT_1
	FIN_WAIT_2
	CLOSE_WAIT
	CLOSING
	LAST_ACK
	TIME_WAIT
	CLOSED
)

func (x *TCP_Client_Manager) New_TCP_Client(src, dst uint16, dstIP string) (*TCP_Client, error) {
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

	return &TCP_Client{
		src: src,
		dst: dst,
		ipAddress: dstIP,
		srcIP: "127.0.0.1", // TODO: don't hardcode the srcIP
		manager: x,
		read: read,
		writer: r,
		seqNum: uint32(3425), // Can be any number between 0 and 2^32 inclusive, TODO this needs to be changed later, it can't be predictable.
		ackNum: uint32(0),    // Always 0 at start
	}, nil
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

func (c *TCP_Client) Connect() error {
	// SYN
	window := uint16(43690)       // TODO calc using http://ithitman.blogspot.com/2013/02/understanding-tcp-window-window-scaling.html
	headerLenSYN := 40
	SYN := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte
		(byte)(c.dst >> 8), (byte)(c.dst), // Destination port in byte slice
		(byte)(c.seqNum >> 24), (byte)(c.seqNum >> 16), (byte)(c.seqNum >> 8), (byte)(c.seqNum),
		(byte)(c.ackNum >> 24), (byte)(c.ackNum >> 16), (byte)(c.ackNum >> 8), (byte)(c.ackNum),
		(byte)(
			(headerLenSYN/4) << 4, // Size of header in 32 bit chunks. It is always 5 unless options are used. This is also the data offset.
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
		0, 0, // checksum
		0, 0, // URG pointer, only matters where URG flag is set.
		0x02, 0x04, 0xff, 0xd7, 0x04, 0x02, 0x08, 0x0a, 0x02, 0x64, 0x80, 0x8b, 0x0, 0x0, 0x0, 0x0, 0x01, 0x03, 0x03, 0x07,
	}

	checksumSYN := checksum(append(append(append(SYN, net.ParseIP(/*c.writer.src*/c.srcIP)...), net.ParseIP(/*c.writer.dst*/c.ipAddress)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(headerLenSYN >> 8), byte(headerLenSYN)}...))
	SYN[16] = byte(checksumSYN >> 8)
	SYN[17] = byte(checksumSYN)

	//c.writer.WriteTo(SYN)
	err := c.writer.WriteTo(&ipv4.Header{
		Version:  ipv4.Version, // protocol version
		Len:      20, // header length
		TOS:      0, // type-of-service (0 is everything normal)
		TotalLen: len(SYN) + 20, // packet total length (octets)
		ID:       0, // identification
		Flags:    ipv4.DontFragment, // flags
		FragOff:  0, // fragment offset
		TTL:      64, // time-to-live (maximum lifespan in seconds)
		Protocol: TCP_PROTO, // next protocol
		Checksum: 0, // checksum (apparently autocomputed)
		Dst: net.ParseIP(c.ipAddress), // destination address
		//Options                         // options, extension headers
	}, SYN, nil)
	fmt.Println("Sent SYN")
	if err != nil {
		fmt.Println("Raw conn send", err)
	}

	// TODO: resend syn if needed

	// SYN-ACK
	fmt.Println("Waiting for syn-ack")
	synack := <- c.read
	fmt.Println("Syn-Ack Rcvd:", synack)

	// ACK
	headerLenACK := 20
	// TODO: verify the syn-ack flags (seq, ack, options-window scale, etc), and checksum
	c.seqNum++ // A+1
	B := uint32(synack[4]) << 24 | uint32(synack[5]) << 16 | uint32(synack[6]) << 8 | uint32(synack[7])
	c.ackNum = B+1

	ACK := []byte{
		(byte)(c.src >> 8), (byte)(c.src), // Source port in byte
		(byte)(c.dst >> 8), (byte)(c.dst), // Destination port in byte slice
		(byte)(c.seqNum >> 24), (byte)(c.seqNum >> 16), (byte)(c.seqNum >> 8), (byte)(c.seqNum),
		(byte)(c.ackNum >> 24), (byte)(c.ackNum >> 16), (byte)(c.ackNum >> 8), (byte)(c.ackNum),
		byte((headerLenACK/4) << 4), // data offset
		byte(TCP_ACK), // flags
		(byte)(window >> 8), (byte)(window), // window size
		0, 0, // checksum
		0, 0, // urg
	}

	checksumACK := checksum(append(append(append(ACK, net.ParseIP(c.srcIP)...), net.ParseIP(c.ipAddress)...), []byte{byte(TCP_PROTO >> 8), byte(TCP_PROTO), byte(headerLenACK >> 8), byte(headerLenACK)}...))
	ACK[16] = byte(checksumACK >> 8)
	ACK[17] = byte(checksumACK)

	err = c.writer.WriteTo(&ipv4.Header{
		Version:  ipv4.Version, // protocol version
		Len:      20, // header length
		TOS:      0, // type-of-service (0 is everything normal)
		TotalLen: len(ACK) + 20, // packet total length (octets)
		ID:       0, // identification
		Flags:    ipv4.DontFragment, // flags
		FragOff:  0, // fragment offset
		TTL:      64, // time-to-live (maximum lifespan in seconds)
		Protocol: TCP_PROTO, // next protocol
		Checksum: 0, // checksum (apparently autocomputed)
		Dst: net.ParseIP(c.ipAddress), // destination address
		//Options                         // options, extension headers
	}, ACK, nil)
	fmt.Println("Sent ACK data")
	if err != nil {
		fmt.Println("Raw conn send", err)
	}

	return nil
}

func (c *TCP_Client) Send(data []byte) error {
	return nil // TODO: implement TCP_Client send
}

func (c *TCP_Client) Close() error {
	return nil // TODO: free manager read buffer
}
