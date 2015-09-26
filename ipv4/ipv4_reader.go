package ipv4

import (
	"network/ipv4/ipv4tps"

	"github.com/hsheth2/logs"

	"sync"
	"errors"
)

type IP_Read_Header struct {
	Rip, Lip *ipv4tps.IPaddress
	B, Payload []byte
}

type ipv4_reader struct {
	incomingPackets chan []byte
	processed chan *IP_Read_Header
	irm             *IP_Read_Manager
	protocol        uint8
	ip              *ipv4tps.IPaddress
	fragBuf         map[string](chan []byte)
	fragBufMutex    *sync.Mutex
}

func NewIP_Reader(irm *IP_Read_Manager, ip *ipv4tps.IPaddress, protocol uint8) (*ipv4_reader, error) {
	c, err := irm.Bind(ip, protocol)
	if err != nil {
		return nil, err
	}

	ipr := &ipv4_reader{
		incomingPackets: c,
		processed: make(chan *IP_Read_Header, IP_READ_MANAGER_BUFFER_SIZE),
		protocol:        protocol,
		ip:              ip,
		fragBuf:         make(map[string](chan []byte)),
		fragBufMutex:    &sync.Mutex{},
	}

	go ipr.readAll()

	return ipr, nil
}

func slicePacket(b []byte) (hrd, payload []byte) {
	hdrLen := int(b[0]&0x0f) * 4
	//fmt.Println("HdrLen: ", hdrLen)
	return b[:hdrLen], b[hdrLen:]
}

func (ipr *ipv4_reader) readAll() {
	for {
		//fmt.Println("STARTING READ")
		b := <-ipr.incomingPackets
		//	//ch logs.Info.Println("Read IP packet")
		//fmt.Println("RAW READ COMPLETED")
		//fmt.Println("Read Length: ", len(b))
		//fmt.Print(".")
		//fmt.Println("Full Read Data: ", b)
		err := ipr.readOne(b)
		if err != nil {
			logs.Error.Println(err)
			continue
		}
	}
}

func (ipr *ipv4_reader) readOne(b []byte) error {
	hdr, p := slicePacket(b)

	// extract source IP and protocol
	rip := &ipv4tps.IPaddress{IP: hdr[12:16]}
	lip := &ipv4tps.IPaddress{IP: hdr[16:20]}

	// verify checksum
	if !verifyIPChecksum(hdr) {
		return errors.New("Header checksum incorrect, packet dropped")
	}

	packetOffset := uint16(hdr[6]&0x1f)<<8 + uint16(hdr[7])
	//fmt.Println("PACK OFF", packetOffset, "HEADER FLAGS", (hdr[6] >> 5))
	if ((hdr[6]>>5)&0x01 == 0) && (packetOffset == 0) {
		// not a fragment
		packet := &IP_Read_Header{
			Rip: rip,
			Lip: lip,
			B: b,
			Payload: p,
		}
		select {
		case ipr.processed <- packet:
		default:
			return errors.New("Dropping packet: no space in buffer")
		}
		return nil
	} else {
		// is a fragment
		bufID := string([]byte{hdr[12], hdr[13], hdr[14], hdr[15], // the source IP
			hdr[16], hdr[17], hdr[18], hdr[19], // the destination IP
			hdr[9],         // protocol
			hdr[4], hdr[5], // identification
		})
		//Trace.Printf("rcvd a fragment-bufId: %x, len: %d\n", bufID, len(b))

		ipr.fragBufMutex.Lock()
		if _, ok := ipr.fragBuf[bufID]; !ok {
			// create the fragment buffer and quit
			//Trace.Printf("creating a new buffer for %x\n", bufID)
			ipr.fragBuf[bufID] = make(chan []byte, FRAGMENT_ASSEMBLER_BUFFER_SIZE)

			quit := make(chan bool, 1)
			done := make(chan bool, 1)
			didQuit := make(chan bool, 1)

			// create the packet assembler in a goroutine to allow the program to continue
			go ipr.fragAssembler(ipr.fragBuf[bufID], quit, didQuit, done)
			go ipr.killFragAssembler(quit, didQuit, done, bufID)
		}

		// send the packet to the assembler
		select {
		case ipr.fragBuf[bufID] <- b:
		default:
			logs.Warn.Println("Dropping fragmented packet, no space in fragment buffer")
		}
		ipr.fragBufMutex.Unlock()

		// after dealing with the fragment
		return nil
	}
}

func (ipr *ipv4_reader) ReadFrom() (*IP_Read_Header, error) {
	return <-ipr.processed, nil
}

func (ipr *ipv4_reader) Close() error {
	return ipr.irm.Unbind(ipr.ip, ipr.protocol)
}

/* h := &ipv4.Header{
	Version:  ipv4.Version,      // protocol version
	Len:      20,                // header length
	TOS:      0,                 // type-of-service (0 is everything normal)
	TotalLen: len(x) + 20,       // packet total length (octets)
	ID:       0,                 // identification
	Flags:    ipv4.DontFragment, // flags
	FragOff:  0,                 // fragment offset
	TTL:      8,                 // time-to-live (maximum lifespan in seconds)
	Protocol: 17,                // next protocol (17 is UDP)
	Checksum: 0,                 // checksum (apparently autocomputed)
	//Src:    net.IPv4(127, 0, 0, 1), // source address, apparently done automatically
	Dst: net.ParseIP(c.manager.ipAddress), // destination address
	//Options                         // options, extension headers
}
*/
