package ipv4

import (
	"errors"

	"github.com/hsheth2/gonet/ethernet"

	"github.com/hsheth2/logs"
)

type ipReadManager struct {
	read    ethernet.Reader
	buffers map[uint8](map[Hash](chan []byte))
}

var globalIPReadManager = func() *ipReadManager {
	irm, err := newIPReadManager()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return irm
}()

func newIPReadManager() (*ipReadManager, error) {
	r, err := ethernet.Bind(ethernet.EtherTypeIP)
	logs.Info.Printf("bound IP reader to EtherType: %.4x", ethernet.EtherTypeIP)
	if err != nil {
		return nil, err
	}

	irm := &ipReadManager{
		read:    r,
		buffers: make(map[uint8](map[Hash](chan []byte))),
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
		//logs.Info.Println("IP read_manager recvd packet:", ethPacket.Packet)
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
	srcIp := &Address{IP: buf[12:16]}
	dstIp := &Address{IP: buf[16:20]}

	logs.Trace.Printf("IP Proto %.2x, %s -> %s", protocol, srcIp, dstIp)
	if protoBuf, foundProto := irm.buffers[protocol]; foundProto {
		//fmt.Println("Dealing with packet")
		var output chan []byte
		if c, foundIP := protoBuf[srcIp.Hash()]; foundIP {
			//fmt.Println("Found exact")
			output = c
		} else if c, foundAll := protoBuf[IPAllHash]; foundAll {
			//fmt.Println("Found global")
			output = c
		} else {
			logs.Warn.Println("not bound to IP ", dstIp)
			return nil
		}
		select {
		case output <- buf:
			//logs.Trace.Println("IP Read Manager forwarding packet")
		default:
			logs.Warn.Println("Dropping incoming IPv4 packet: no space in buffer")
		}
	} else {
		logs.Trace.Printf("Nothing bound to IP proto %.2x, dropping", protocol)
	}
	return nil
}

func (irm *ipReadManager) bind(ip *Address, protocol uint8) (chan []byte, error) {
	// create the protocol buffer if it doesn't exist already
	_, protoOk := irm.buffers[protocol]
	if !protoOk {
		irm.buffers[protocol] = make(map[Hash](chan []byte))
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

func (irm *ipReadManager) unbind(ip *Address, protocol uint8) error {
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

func Noop() {}
