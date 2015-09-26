package ipv4

import (
	"errors"
	"network/ethernet"
	"network/ipv4/ipv4tps"

	"github.com/hsheth2/logs"
)

const IP_READ_MANAGER_BUFFER_SIZE = 5000

type ip_read_manager struct {
	read    ethernet.Ethernet_Reader
	buffers map[uint8](map[ipv4tps.IPhash](chan []byte))
}

var globalIP_Read_Manager = func() *ip_read_manager {
	irm, err := NewIP_Read_Manager(ethernet.GlobalNetworkReadManager)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return irm
}()

func NewIP_Read_Manager(in *ethernet.Network_Read_Manager) (*ip_read_manager, error) {
	r, err := in.Bind(ethernet.ETHERTYPE_IP)
	if err != nil {
		return nil, err
	}

	irm := &ip_read_manager{
		read:    r,
		buffers: make(map[uint8](map[ipv4tps.IPhash](chan []byte))),
	}

	go irm.readAll()

	return irm, nil
}

func (nr *ip_read_manager) readAll() {
	for {
		//logs.Trace.Println("IP read manager readAll starting")
		eth_packet, err := nr.read.Read()
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//logs.Info.Println("IP read_manager recvd packet:", eth_packet.Packet)
		buf := eth_packet.Packet

		if len(buf) <= IP_HEADER_LEN {
			logs.Warn.Println("Dropping IP Packet for bogus length <=", IP_HEADER_LEN)
			logs.Warn.Println("Data being dropped:", buf)
			continue
		}

		err = nr.processOne(buf)
		if err != nil {
			logs.Warn.Println(err)
			continue
		}
	}
}

func (nr *ip_read_manager) processOne(buf []byte) error {
	protocol := uint8(buf[9])
	rip := &ipv4tps.IPaddress{IP: buf[12:16]}

	//fmt.Println(ln)
	//fmt.Println(protocol, ip)
	/*if ln == 47 {
		fmt.Println(buf)
	}*/
	//fmt.Println(protocol, ip)
	if protoBuf, foundProto := nr.buffers[protocol]; foundProto {
		//fmt.Println("Dealing with packet")
		var output chan []byte = nil
		if c, foundIP := protoBuf[rip.Hash()]; foundIP {
			//fmt.Println("Found exact")
			output = c
		} else if c, foundAll := protoBuf[ipv4tps.IP_ALL_HASH]; foundAll {
			//fmt.Println("Found global")
			output = c
		} else {
			logs.Warn.Println("output buf doesn't exist, rip:", rip)
			return nil
		}
		select {
		case output <- buf:
			//logs.Trace.Println("IP Read Manager forwarding packet")
		default:
			logs.Warn.Println("Dropping incoming IPv4 packet: no space in buffer")
		}
	}
	return nil
}

func (irm *ip_read_manager) Bind(ip *ipv4tps.IPaddress, protocol uint8) (chan []byte, error) {
	// create the protocol buffer if it doesn't exist already
	_, protoOk := irm.buffers[protocol]
	if !protoOk {
		irm.buffers[protocol] = make(map[ipv4tps.IPhash](chan []byte))
		//Trace.Println("Bound to", protocol)
	}

	// add the IP binding, if possible
	if _, IP_exists := irm.buffers[protocol][ip.Hash()]; !IP_exists {
		// doesn't exist in map already
		buf := make(chan []byte, IP_READ_MANAGER_BUFFER_SIZE)
		irm.buffers[protocol][ip.Hash()] = buf
		return buf, nil
	}
	return nil, errors.New("IP already bound to.")
}

func (irm *ip_read_manager) Unbind(ip *ipv4tps.IPaddress, protocol uint8) error {
	ipBuf, protoOk := irm.buffers[protocol]
	if !protoOk {
		return errors.New("IP not bound, cannot unbind")
	}

	if _, ok := ipBuf[ip.Hash()]; ok {
		delete(ipBuf, ip.Hash())
		return nil
	}
	return errors.New("Not bound, can't unbind.")
}
