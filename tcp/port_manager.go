package tcp

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hsheth2/gonet/ipv4"

	"github.com/hsheth2/logs"
)

// Global src, dst port and ip registry for TCP binding
type portManagerType struct {
	tcpReader ipv4.Reader
	incoming  map[uint16](map[uint16](map[ipv4.Hash](chan *packet))) // dst, src port, remote ip
	lock      *sync.RWMutex
}

func (m *portManagerType) GetUnusedPort() (uint16, error) {
	// race prevention
	m.lock.RLock()
	defer m.lock.RUnlock()

	for i := minPort; i <= maxPort; i++ {
		if _, exists := m.incoming[i]; !exists {
			return i, nil
		}
	}
	return uint16(0), errors.New("No ports available to bind to")
}

var portManager *portManagerType

func init() {
	ipr, err := ipv4.NewReader(ipv4.IPAll, ipv4.IPProtoTCP)
	if err != nil {
		logs.Error.Println(err)
		return
	}

	portManager = &portManagerType{
		tcpReader: ipr,
		incoming:  make(map[uint16](map[uint16](map[ipv4.Hash](chan *packet)))),
		lock:      &sync.RWMutex{},
	}
	go portManager.readAll()
}

func (m *portManagerType) bind(rport, lport uint16, ip *ipv4.Address) (chan *packet, error) {
	// race prevention
	m.lock.Lock()
	defer m.lock.Unlock()

	// lport is the local one here, rport is the remote
	///*logs*/ logs.Info.Println("Attempting to bind to rport", rport, "lport", lport, "ip", ip.Hash())
	if _, ok := m.incoming[lport]; !ok {
		m.incoming[lport] = make(map[uint16](map[ipv4.Hash](chan *packet)))
	}

	// TODO add an option (for servers) for all srcports
	if _, ok := m.incoming[lport][rport]; !ok {
		m.incoming[lport][rport] = make(map[ipv4.Hash](chan *packet))
	}

	if _, ok := m.incoming[lport][rport][ip.Hash()]; ok {
		return nil, fmt.Errorf("Ports (lport: %d, rport %d) and IP (%v) already bound", lport, rport, ip)
	}

	ans := make(chan *packet, incomingBufferSize)
	m.incoming[lport][rport][ip.Hash()] = ans
	return ans, nil
}

func (m *portManagerType) unbind(rport, lport uint16, ip *ipv4.Address) error {
	// race prevention
	m.lock.Lock()
	defer m.lock.Unlock()

	// TODO verify that it actually won't crash
	close(m.incoming[lport][rport][ip.Hash()])
	/*logs*/ logs.Trace.Println("Closing the packetdealer channel")
	delete(m.incoming[lport][rport], ip.Hash())
	return nil
}

func (m *portManagerType) readAll() {
	for {
		header, err := m.tcpReader.ReadFrom()
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

func (m *portManagerType) readDeal(rip, lip *ipv4.Address, payload []byte) error {
	p, err := extractPacket(payload, rip, lip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	rport := p.header.srcport
	lport := p.header.dstport

	var output chan *packet

	m.lock.RLock()
	defer m.lock.RUnlock()
	///*logs*/logs.Trace.Printf("readAll tcp packet manager dealing with packet or rport: %d and lport %d", rport, lport)
	if _, ok := m.incoming[lport]; ok {
		///*logs*/logs.Trace.Printf("readAll: promising packet rport: %d and lport %d", rport, lport)
		if p, ok := m.incoming[lport][rport]; ok {
			///*logs*/logs.Trace.Println("readAll: exact port number match")
			if x, ok := p[rip.Hash()]; ok {
				output = x
			} else if x, ok := p[ipv4.IPAllHash]; ok {
				output = x
			}
		} else if p, ok := m.incoming[lport][0]; ok {
			///*logs*/logs.Trace.Println("readAll: forwarding to a listening server")
			if x, ok := p[ipv4.IPAllHash]; ok {
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
