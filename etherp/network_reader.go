package etherp

import (
	"errors"
	"network/logs"
	"net"
	"syscall"
)

/*type Network_Reader_IP struct {
    dst string
    protocol uint8
    buffer chan []byte
}

func NewNetwork_Reader_IP(dst string, protocol uint8) (*Network_Reader_IP, error) {

}*/

var GlobalNetworkReader = func() *Network_Reader {
	x, err := NewNetwork_Reader()
	if err != nil {
		logs.Error.Fatal(err)
	}
	return x
}()

type Network_Reader struct {
	fd      int
	buffers map[uint8](map[string](chan []byte))
	//buffers map[uint8](chan []byte)
}

func NewNetwork_Reader() (*Network_Reader, error) {
	fd, err := syscall.Socket(AF_PACKET, SOCK_RAW, HTONS_ETH_P_ALL)

	if err != nil {
		logs.Error.Println("AF_PACKET socket connection")
		return nil, err
	}

	nr := &Network_Reader{
		fd:      fd,
		buffers: make(map[uint8](map[string](chan []byte))),
	}
	go nr.readAll()

	return nr, nil
}

func (nr *Network_Reader) readAll() {
	for {
		// read twice to account for the double receiving TODO fix the double reading somehow
		buf := make([]byte, MAX_IP_PACKET_LEN)
		_, err := nr.getNextPacket(buf)
		ln, err := nr.getNextPacket(buf)

		if err != nil {
			logs.Error.Println(err)
		}
		buf = buf[:ln] // remove extra bytes off the end

		//fmt.Println("Ethernet header:", buf[:14])
		// TODO: verify the ethernet protocol legitimately
		eth_protocol := uint16(buf[12])<<8 | uint16(buf[13])
		if eth_protocol != ETHERTYPE_IP { // verify that protocol is 0x0800 for IP
			//Info.Println("Dropping Ethernet packet for wrong protocol:", eth_protocol)
			continue
		}
		buf = buf[14:] // remove ethernet header
		//fmt.Println("After removing ethernet header", buf)

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

func (nr *Network_Reader) Bind(ip string, protocol uint8) (chan []byte, error) {
	// create the protocol buffer if it doesn't exist already
	_, protoOk := nr.buffers[protocol]
	if !protoOk {
		nr.buffers[protocol] = make(map[string](chan []byte))
		//Trace.Println("Bound to", protocol)
	}

	// add the IP binding, if possible
	if _, IP_exists := nr.buffers[protocol][ip]; !IP_exists {
		// doesn't exist in map already
		nr.buffers[protocol][ip] = make(chan []byte, 1)

		ret, _ := nr.buffers[protocol][ip]
		return ret, nil
	}
	return nil, errors.New("IP already bound to.")
}

func (nr *Network_Reader) Unbind(ip string, protocol uint8) error {
	ipBuf, protoOk := nr.buffers[protocol]
	if !protoOk {
		return errors.New("IP not bound, cannot unbind")
	}

	if _, ok := ipBuf[ip]; ok {
		delete(ipBuf, ip)
		return nil
	}
	return errors.New("Not bound, can't unbind.")
}

func (nr *Network_Reader) getNextPacket(buf []byte) (int, error) {
	return syscall.Read(nr.fd, buf)
}
