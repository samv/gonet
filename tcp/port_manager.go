package tcp

import (
	"network/ipv4"

	"github.com/hsheth2/logs"

	"network/ipv4/ipv4tps"

	"sync"

	"fmt"
)

// Global src, dst port and ip registry for TCP binding
type TCP_Port_Manager_Type struct {
	tcp_reader ipv4.IPv4_Reader
	incoming   map[uint16](map[uint16](map[ipv4tps.IPhash](chan *TCP_Packet))) // dst, src port, remote ip
	lock       *sync.RWMutex
}

func (m *TCP_Port_Manager_Type) bind(rport, lport uint16, ip *ipv4tps.IPaddress) (chan *TCP_Packet, error) {
	// race prevention
	m.lock.Lock()
	defer m.lock.Unlock()

	// lport is the local one here, rport is the remote
	//ch logs.Info.Println("Attempting to bind to rport", rport, "lport", lport, "ip", ip.Hash())
	if _, ok := m.incoming[lport]; !ok {
		m.incoming[lport] = make(map[uint16](map[ipv4tps.IPhash](chan *TCP_Packet)))
	}

	// TODO add an option (for servers) for all srcports
	if _, ok := m.incoming[lport][rport]; !ok {
		m.incoming[lport][rport] = make(map[ipv4tps.IPhash](chan *TCP_Packet))
	}

	if _, ok := m.incoming[lport][rport][ip.Hash()]; ok {
		return nil, fmt.Errorf("Ports (lport: %d, rport %d) and IP (%v) already binded to", lport, rport, ip)
	}

	ans := make(chan *TCP_Packet, TCP_INCOMING_BUFF_SZ)
	m.incoming[lport][rport][ip.Hash()] = ans
	return ans, nil
}

func (m *TCP_Port_Manager_Type) unbind(rport, lport uint16, ip *ipv4tps.IPaddress) error {
	// race prevention
	m.lock.Lock()
	defer m.lock.Unlock()

	// TODO verify that it actually won't crash
	close(m.incoming[lport][rport][ip.Hash()])
	//ch logs.Trace.Println("Closing the packetdealer channel")
	delete(m.incoming[lport][rport], ip.Hash())
	return nil
}

func (m *TCP_Port_Manager_Type) readAll() {
	for {
		header, err := m.tcp_reader.ReadFrom()
		if err != nil {
			logs.Error.Println("TCP readAll error", err)
			continue
		}

		err = m.readDeal(header.Rip, header.Lip, header.Payload)
		if err != nil {
			logs.Error.Println(err)
			continue
		}
	}
}

func (m *TCP_Port_Manager_Type) readDeal(rip, lip *ipv4tps.IPaddress, payload []byte) error {
	p, err := Extract_TCP_Packet(payload, rip, lip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	rport := p.header.srcport
	lport := p.header.dstport

	var output chan *TCP_Packet = nil

	m.lock.RLock()
	defer m.lock.RUnlock()
	////ch logs.Trace.Printf("readAll tcp packet manager dealing with packet or rport: %d and lport %d", rport, lport)
	if _, ok := m.incoming[lport]; ok {
		////ch logs.Trace.Printf("readAll: promising packet rport: %d and lport %d", rport, lport)
		if p, ok := m.incoming[lport][rport]; ok {
			////ch logs.Trace.Println("readAll: exact port number match")
			if x, ok := p[rip.Hash()]; ok {
				output = x
			} else if x, ok := p[ipv4tps.IP_ALL_HASH]; ok {
				output = x
			}
		} else if p, ok := m.incoming[lport][0]; ok {
			////ch logs.Trace.Println("readAll: forwarding to a listening server")
			if x, ok := p[ipv4tps.IP_ALL_HASH]; ok {
				output = x
			} else if x, ok := p[rip.Hash()]; ok {
				output = x
			}
		}
	}

	if output != nil {
		select {
		case output <- p:
		default:
			logs.Warn.Println("Dropping TCP packet: no space in buffer")
		}
	} else {
		// TODO send a rst to sender if nothing is binded to the dst port, src port, and remote ip
		//fmt.Println(errors.New("Dst/Src port + ip not binded to"))
		logs.Warn.Println(fmt.Errorf("Dropping TCP packet (lport: %d, rport %d); nothing listening for it", lport, rport))
	}

	return nil
}

var TCP_Port_Manager = func() *TCP_Port_Manager_Type {
	ipr, err := ipv4.NewIP_Reader(ipv4tps.IP_ALL, ipv4.TCP_PROTO)
	if err != nil {
		logs.Error.Println(err)
		return nil
	}

	m := &TCP_Port_Manager_Type{
		tcp_reader: ipr,
		incoming:   make(map[uint16](map[uint16](map[ipv4tps.IPhash](chan *TCP_Packet)))),
		lock:       &sync.RWMutex{},
	}
	go m.readAll()
	return m
}()
