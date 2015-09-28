package ipv4

import (
	"errors"
	"network/ethernet"

	"github.com/hsheth2/logs"
)

type ipReadManager struct {
	read    ethernet.Reader
	buffers map[uint8](map[IPhash](chan []byte))
}

var globalIPReadManager = func() *ipReadManager {
	irm, err := newIPReadManager(ethernet.GlobalNetworkReadManager)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return irm
}()

func newIPReadManager(in *ethernet.NetworkReadManager) (*ipReadManager, error) {
	r, err := in.Bind(ethernet.EtherTypeIP)
	if err != nil {
		return nil, err
	}

	irm := &ipReadManager{
		read:    r,
		buffers: make(map[uint8](map[IPhash](chan []byte))),
	}

	go irm.readAll()

	return irm, nil
}

func (irm *ipReadManager) readAll() {
	for {
		//logs.Trace.Println("IP read manager readAll starting")
		ethPacket, err := irm.read.Read()
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//logs.Info.Println("IP read_manager recvd packet:", eth_packet.Packet)
		buf := ethPacket.Packet

		if len(buf) <= ipHeaderLength {
			logs.Warn.Println("Dropping IP Packet for bogus length <=", ipHeaderLength)
			logs.Warn.Println("Data being dropped:", buf)
			continue
		}

		err = irm.processOne(buf)
		if err != nil {
			logs.Warn.Println(err)
			continue
		}
	}
}

func (irm *ipReadManager) processOne(buf []byte) error {
	protocol := uint8(buf[9])
	rip := &IPAddress{IP: buf[12:16]}

	//fmt.Println(ln)
	//fmt.Println(protocol, ip)
	/*if ln == 47 {
		fmt.Println(buf)
	}*/
	//fmt.Println(protocol, ip)
	if protoBuf, foundProto := irm.buffers[protocol]; foundProto {
		//fmt.Println("Dealing with packet")
		var output chan []byte
		if c, foundIP := protoBuf[rip.Hash()]; foundIP {
			//fmt.Println("Found exact")
			output = c
		} else if c, foundAll := protoBuf[IPAllHash]; foundAll {
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

func (irm *ipReadManager) bind(ip *IPAddress, protocol uint8) (chan []byte, error) {
	// create the protocol buffer if it doesn't exist already
	_, protoOk := irm.buffers[protocol]
	if !protoOk {
		irm.buffers[protocol] = make(map[IPhash](chan []byte))
		//Trace.Println("Bound to", protocol)
	}

	// add the IP binding, if possible
	if _, exists := irm.buffers[protocol][ip.Hash()]; !exists {
		// doesn't exist in map already
		buf := make(chan []byte, ipReadBufferSize)
		irm.buffers[protocol][ip.Hash()] = buf
		return buf, nil
	}
	return nil, errors.New("IP already bound to")
}

func (irm *ipReadManager) unbind(ip *IPAddress, protocol uint8) error {
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
