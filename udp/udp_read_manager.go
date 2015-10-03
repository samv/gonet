package udp

import (
	"errors"
	"fmt"
	"network/ipv4"

	"github.com/hsheth2/logs"
)

type UDP_Read_Manager struct {
	reader ipv4.Reader
	buff   map[uint16](map[ipv4.Hash](chan []byte))
}

var GlobalUDP_Read_Manager *UDP_Read_Manager = func() *UDP_Read_Manager {
	rm, err := NewUDP_Read_Manager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return rm
}()

func NewUDP_Read_Manager() (*UDP_Read_Manager, error) {
	ipr, err := ipv4.NewReader(ipv4.IPAll, ipv4.IPProtoUDP)
	if err != nil {
		return nil, err
	}

	x := &UDP_Read_Manager{
		reader: ipr,
		buff:   make(map[uint16](map[ipv4.Hash](chan []byte))),
	}

	go x.readAll()

	return x, nil
}

func (x *UDP_Read_Manager) readAll() {
	for {
		header, err := x.reader.ReadFrom()
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//fmt.Println(b)
		//fmt.Println("UDP header and payload: ", payload)

		dst := (((uint16)(header.Payload[2])) * 256) + ((uint16)(header.Payload[3]))
		//fmt.Println(dst)

		if len(header.Payload) < UDP_HEADER_SZ {
			//ch logs.Info.Println("Dropping Small UDP packet:", payload)
			continue
		}

		headerLen := uint16(header.Payload[4])<<8 | uint16(header.Payload[5])
		if !ipv4.VerifyTransportChecksum(header.Payload[:UDP_HEADER_SZ], header.Rip, header.Lip, headerLen, ipv4.IPProtoUDP) {
			//ch logs.Info.Println("Dropping UDP Packet for bad checksum:", payload)
			continue
		}

		header.Payload = header.Payload[UDP_HEADER_SZ:]
		//fmt.Println(payload)

		if portBuf, ok := x.buff[dst]; ok {
			var output chan []byte = nil
			if c, ok := portBuf[header.Rip.Hash()]; ok {
				//fmt.Println("Found exact IP match for port", dst)
				output = c
			} else if c, ok := portBuf[ipv4.IPAllHash]; ok {
				//fmt.Println("Found default IP match for port", dst)
				output = c
			} else {
				logs.Warn.Println("Dropping UDP packet because nothing wanted it")
			}
			select {
			case output <- header.Payload:
				//ch logs.Trace.Println("Forwarded UDP packet lport:", dst, "and rip:", rip.IP)
			default:
				logs.Warn.Println("Dropping UDP packet: no space in buffer")
			}
		} else {
			////ch logs.Info.Println("Dropping UDP packet:", payload)
		}
	}
}

func (x *UDP_Read_Manager) Bind(port uint16, ip *ipv4.Address) (chan []byte, error) {
	// add the port if not already there
	if _, found := x.buff[port]; !found {
		x.buff[port] = make(map[ipv4.Hash](chan []byte))
	}

	// add the ip to the port's list
	if _, found := x.buff[port][ip.Hash()]; !found {
		ans := make(chan []byte, UDP_RCV_BUF_SZ)
		x.buff[port][ip.Hash()] = ans
		return ans, nil
	} else {
		return nil, errors.New("Another application is already listening to port " + fmt.Sprintf("%v", port) + " with IP " + fmt.Sprintf("%v", ip))
	}
}

func (x *UDP_Read_Manager) Unbind(port uint16, ip *ipv4.Address) error {
	delete(x.buff[port], ip.Hash()) // TODO verify that it will succeed
	return nil
}
