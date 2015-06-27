package ipv4p

import (
	"errors"
	"github.com/hsheth2/logs"
	"net"
	"network/etherp"
)

type IP_Read_Manager struct {
	incoming chan []byte
	buffers  map[uint8](map[string](chan []byte))
}

var GlobalIPReadManager = func() *IP_Read_Manager {
	irm, err := NewIP_Read_Manager(etherp.GlobalNetworkReader)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return irm
}()

func NewIP_Read_Manager(in *etherp.Network_Reader) (*IP_Read_Manager, error) {
	input, err := in.Bind(etherp.ETHERTYPE_IP)
	if err != nil {
		return nil, err
	}

	irm := &IP_Read_Manager{
		incoming: input,
		buffers:  make(map[uint8](map[string](chan []byte))),
	}

	go irm.readAll()

	return irm, nil
}

func (nr *IP_Read_Manager) readAll() {
	for {
		buf := <-nr.incoming

		if len(buf) <= IP_HEADER_LEN {
			logs.Info.Println("Dropping IP Packet for bogus length <=", IP_HEADER_LEN)
			continue
		}

		protocol := uint8(buf[9])
		ip := net.IPv4(buf[12], buf[13], buf[14], buf[15]).String()

		//fmt.Println(ln)
		//fmt.Println(protocol, ip)
		/*if ln == 47 {
			fmt.Println(buf)
		}*/
		//fmt.Println(protocol, ip)
		if protoBuf, foundProto := nr.buffers[protocol]; foundProto {
			//fmt.Println("Dealing with packet")
			if c, foundIP := protoBuf[ip]; foundIP {
				//fmt.Println("Found exact")
				go func() { c <- buf }()
			} else if c, foundAll := protoBuf["*"]; foundAll {
				//fmt.Println("Found global")
				go func() { c <- buf }()
			}
		}
	}
}

func (irm *IP_Read_Manager) Bind(ip string, protocol uint8) (chan []byte, error) {
	// create the protocol buffer if it doesn't exist already
	_, protoOk := irm.buffers[protocol]
	if !protoOk {
		irm.buffers[protocol] = make(map[string](chan []byte))
		//Trace.Println("Bound to", protocol)
	}

	// add the IP binding, if possible
	if _, IP_exists := irm.buffers[protocol][ip]; !IP_exists {
		// doesn't exist in map already
		irm.buffers[protocol][ip] = make(chan []byte, 1)

		ret, _ := irm.buffers[protocol][ip]
		return ret, nil
	}
	return nil, errors.New("IP already bound to.")
}

func (irm *IP_Read_Manager) Unbind(ip string, protocol uint8) error {
	ipBuf, protoOk := irm.buffers[protocol]
	if !protoOk {
		return errors.New("IP not bound, cannot unbind")
	}

	if _, ok := ipBuf[ip]; ok {
		delete(ipBuf, ip)
		return nil
	}
	return errors.New("Not bound, can't unbind.")
}
