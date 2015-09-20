package udp

import (
	"errors"
	"fmt"
	"network/ipv4"
	"network/ipv4/ipv4tps"

	"github.com/hsheth2/logs"
)

type UDP_Read_Manager struct {
	reader *ipv4.IP_Reader
	buff   map[uint16](map[ipv4tps.IPhash](chan []byte))
}

var GlobalUDP_Read_Manager *UDP_Read_Manager = func() *UDP_Read_Manager {
	rm, err := NewUDP_Read_Manager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return rm
}()

func NewUDP_Read_Manager() (*UDP_Read_Manager, error) {
	irm := ipv4.GlobalIPReadManager

	ipr, err := ipv4.NewIP_Reader(irm, ipv4tps.IP_ALL, ipv4.UDP_PROTO)
	if err != nil {
		return nil, err
	}

	x := &UDP_Read_Manager{
		reader: ipr,
		buff:   make(map[uint16](map[ipv4tps.IPhash](chan []byte))),
	}

	go x.readAll()

	return x, nil
}

func (x *UDP_Read_Manager) readAll() {
	for {
		rip, lip, _, payload, err := x.reader.ReadFrom()
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//fmt.Println(b)
		//fmt.Println("UDP header and payload: ", payload)

		dst := (((uint16)(payload[2])) * 256) + ((uint16)(payload[3]))
		//fmt.Println(dst)

		if len(payload) < UDP_HEADER_SZ {
			//ch logs.Info.Println("Dropping Small UDP packet:", payload)
			continue
		}

		headerLen := uint16(payload[4])<<8 | uint16(payload[5])
		if !ipv4.VerifyTransportChecksum(payload[:UDP_HEADER_SZ], rip, lip, headerLen, ipv4.UDP_PROTO) {
			//ch logs.Info.Println("Dropping UDP Packet for bad checksum:", payload)
			continue
		}

		payload = payload[UDP_HEADER_SZ:]
		//fmt.Println(payload)

		if portBuf, ok := x.buff[dst]; ok {
			var output chan []byte = nil
			if c, ok := portBuf[rip.Hash()]; ok {
				//fmt.Println("Found exact IP match for port", dst)
				output = c
			} else if c, ok := portBuf[ipv4tps.IP_ALL_HASH]; ok {
				//fmt.Println("Found default IP match for port", dst)
				output = c
			} else {
				logs.Warn.Println("Dropping UDP packet because nothing wanted it")
			}
			select {
			case output <- payload:
				//ch logs.Trace.Println("Forwarded UDP packet lport:", dst, "and rip:", rip.IP)
			default:
				logs.Warn.Println("Dropping UDP packet: no space in buffer")
			}
		} else {
			////ch logs.Info.Println("Dropping UDP packet:", payload)
		}
	}
}

func (x *UDP_Read_Manager) Bind(port uint16, ip *ipv4tps.IPaddress) (chan []byte, error) {
	// add the port if not already there
	if _, found := x.buff[port]; !found {
		x.buff[port] = make(map[ipv4tps.IPhash](chan []byte))
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

func (x *UDP_Read_Manager) Unbind(port uint16, ip *ipv4tps.IPaddress) error {
	delete(x.buff[port], ip.Hash()) // TODO verify that it will succeed
	return nil
}
