package main

import (
	"errors"
	"fmt"
)

const MAX_UDP_PACKET_LEN = 65507

type UDP_Read_Manager struct {
	reader *IP_Reader
	buff   map[uint16](map[string](chan []byte))
}

type UDP_Reader struct {
	manager   *UDP_Read_Manager
	bytes     <-chan []byte
	port      uint16 // ports
	ipAddress string
}

func NewUDP_Read_Manager() (*UDP_Read_Manager, error) {
	nr, err := NewNetwork_Reader()
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader("*", UDP_PROTO) // 17 for UDP
	if err != nil {
		return nil, err
	}

	x := &UDP_Read_Manager{
		reader: ipr,
		buff:   make(map[uint16](map[string](chan []byte))),
	}

	go x.readAll()

	return x, nil
}

func (x *UDP_Read_Manager) readAll() {
	for {
		ip, _, payload, err := x.reader.ReadFrom()
		if err != nil {

			fmt.Println(err)
			continue
		}
		//fmt.Println(b)
		//fmt.Println("UDP header and payload: ", payload)

		dst := (((uint16)(payload[2])) * 256) + ((uint16)(payload[3]))
		//fmt.Println(dst)

		payload = payload[8:]
		//fmt.Println(payload)

		portBuf, ok := x.buff[dst]
		//fmt.Println(ok)
		if ok {
			if c, ok := portBuf[ip]; ok {
				fmt.Println("Found exact IP match for port", dst)
				go func() { c <- payload }()
			} else if c, ok := portBuf["*"]; ok {
				//fmt.Println("Found default IP match for port", dst)
				go func() { c <- payload }()
			}
		} else {
			// drop packet
		}
	}
}

func (x *UDP_Read_Manager) NewUDP(port uint16, ip string) (*UDP_Reader, error) {
	// add the port if not already there
	if _, found := x.buff[port]; !found {
		x.buff[port] = make(map[string](chan []byte))
	}

	// add the ip to the port's list
	if _, found := x.buff[port][ip]; !found {
		x.buff[port][ip] = make(chan []byte)
		return &UDP_Reader{port: port, bytes: x.buff[port][ip], manager: x, ipAddress: ip}, nil
	} else {
		return nil, errors.New("Another application is already listening to that port.")
	}
}

func (c *UDP_Reader) read(size int) ([]byte, error) {
	data := <-c.bytes
	if len(data) > size {
		data = data[:size]
	}
    // TODO: verify the checksum
	return data, nil
}

func (c *UDP_Reader) close() error {
	delete(c.manager.buff, c.port)
	return nil
}
