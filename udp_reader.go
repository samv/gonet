package main

import (
//    "fmt"
)

type UDP_Read_Manager struct {
	ipAddress string
	reader    *IP_Reader
	buff      map[uint16](chan []byte)
}

type UDP_Reader struct {
	manager *UDP_Read_Manager
	bytes   <-chan []byte
	port    uint16 // ports
}

func NewUDP_Read_Manager(ip string) (*UDP_Read_Manager, error) {
	nr, err := NewNetwork_Reader()
	if err != nil {
		return nil, err
	}

	ipr, err := nr.NewIP_Reader(ip, 17) // 17 for UDP
	if err != nil {
		return nil, err
	}

	x := &UDP_Read_Manager{
		reader:    ipr,
		buff:      make(map[uint16](chan []byte)),
		ipAddress: ip,
	}

	go x.readAll()

	return x, nil
}

func (x *UDP_Read_Manager) readAll() {
	for {
		_, payload, err := x.reader.ReadFrom()
		if err != nil {
			continue
		}
		//fmt.Println(b)
		//fmt.Println("UDP header and payload: ", payload)

		dest := (((uint16)(payload[2])) * 256) + ((uint16)(payload[3]))
		//fmt.Println(dest)
		//fmt.Println(payload)
		//
		//fmt.Println(x.buff)
		c, ok := x.buff[dest]
		//fmt.Println(ok)
		payload = payload[8:]
		if ok {
			go func() { c <- payload }()
		} else {
			// drop packet
		}
	}
}

func (x *UDP_Read_Manager) NewUDP(port uint16) (*UDP_Reader, error) {
	x.buff[port] = make(chan []byte)
	return &UDP_Reader{port: port, bytes: x.buff[port], manager: x}, nil
}

func (c *UDP_Reader) read(size int) ([]byte, error) {
	data := <-c.bytes
	if len(data) > size {
		data = data[:size]
	}
	return data, nil
}

func (c *UDP_Reader) close() error {
	delete(c.manager.buff, c.port)
	return nil
}
