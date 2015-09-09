package ipv4

import (
	"network/ipv4/ipv4tps"
	"time"

	"github.com/hsheth2/logs"
)

type IP_Reader struct {
	incomingPackets chan []byte
	irm             *IP_Read_Manager
	protocol        uint8
	ip              *ipv4tps.IPaddress
	fragBuf         map[string](chan []byte)
}

func NewIP_Reader(irm *IP_Read_Manager, ip *ipv4tps.IPaddress, protocol uint8) (*IP_Reader, error) {
	c, err := irm.Bind(ip, protocol)
	if err != nil {
		return nil, err
	}

	return &IP_Reader{
		incomingPackets: c,
		protocol:        protocol,
		ip:              ip,
		fragBuf:         make(map[string](chan []byte)),
	}, nil
}

func slicePacket(b []byte) (hrd, payload []byte) {
	hdrLen := int(b[0]&0x0f) * 4
	//fmt.Println("HdrLen: ", hdrLen)
	return b[:hdrLen], b[hdrLen:]
}

func (ipr *IP_Reader) fragmentAssembler(in <-chan []byte, quit <-chan bool, didQuit chan<- bool, done chan bool) {
	payload := make([]byte, 0)
	extraFrags := make(map[uint64]([]byte))
	recvLast := false

	for {
		select {
		case <-quit:
			//Trace.Println("quitting upon quit signal")
			didQuit <- true
			return
		case frag := <-in:
			//Trace.Println("got a fragment packet. len:", len(frag))
			hdr, p := slicePacket(frag)
			//offset := 8 * (uint64(hdr[6]&0x1f)<<8 + uint64(hdr[7]))
			//fmt.Println("RECEIVED FRAG")
			//fmt.Println("Offset:", offset)
			//fmt.Println(len(payload))

			// add to map
			offset := 8 * (uint64(hdr[6]&0x1F)<<8 + uint64(hdr[7]))
			//Trace.Println("Offset:", offset)
			extraFrags[offset] = p

			// check more fragments flag
			if (hdr[6]>>5)&0x01 == 0 {
				recvLast = true
			}

			// add to payload
			//Trace.Println("Begin to add to the payload")
			for {
				if storedFrag, found := extraFrags[uint64(len(payload))]; found {
					//Trace.Println("New Payload Len: ", len(payload))
					delete(extraFrags, uint64(len(payload)))
					payload = append(payload, storedFrag...)
				} else {
					break
				}
			}
			//Trace.Println("Finished add to the payload")

			// deal with the payload
			if recvLast && len(extraFrags) == 0 {
				//Trace.Println("Done")
				// correct the header
				fullPacketHdr := hdr
				totalLen := uint16(fullPacketHdr[0]&0x0F)*4 + uint16(len(payload))
				fullPacketHdr[2] = byte(totalLen >> 8)
				fullPacketHdr[3] = byte(totalLen)
				fullPacketHdr[6] = 0
				fullPacketHdr[7] = 0

				// update the checksum
				check := calculateIPChecksum(fullPacketHdr[:20])
				fullPacketHdr[10] = byte(check >> 8)
				fullPacketHdr[11] = byte(check)

				// send the packet back into processing
				//go func() {
				select {
				case ipr.incomingPackets <- append(fullPacketHdr, payload...):
				default:
					logs.Warn.Println("Dropping defragmented packet, no space in buffer")
				}
				//fmt.Println("FINISHED")
				//}()
				//Trace.Println("Just wrote back in")
				done <- true
				return // from goroutine
			}
			//Trace.Println("Looping")
		}
	}

	// drop the packet upon timeout
	//Trace.Println(errors.New("Fragments took too long, packet dropped"))
	//return
}

func (ipr *IP_Reader) killFragmentAssembler(quit chan<- bool, didQuit <-chan bool, done <-chan bool, bufID string) {
	// sends quit to the assembler if it doesn't send done
	select {
	case <-time.After(FRAGMENT_TIMEOUT):
		//Trace.Println("Force quitting packet assembler")
		quit <- true
		<-didQuit // will block until it has been received
	case <-done:
		//Trace.Println("Received done msg.")
	}

	//Trace.Println("Frag Assemble Ended, finished")
	delete(ipr.fragBuf, bufID)
}

func (ipr *IP_Reader) ReadFrom() (rip, lip *ipv4tps.IPaddress, b, payload []byte, e error) {
	//fmt.Println("STARTING READ")
	b = <-ipr.incomingPackets
	//	logs.Info.Println("Read IP packet")
	//fmt.Println("RAW READ COMPLETED")
	//fmt.Println("Read Length: ", len(b))
	//fmt.Print(".")
	//fmt.Println("Full Read Data: ", b)

	hdr, p := slicePacket(b)

	// extract source IP and protocol
	rip = &ipv4tps.IPaddress{IP: hdr[12:16]}
	lip = &ipv4tps.IPaddress{IP: hdr[16:20]}
	//	proto := uint8(hdr[9])
	//	if !((bytes.Equal(ipr.ip.IP, rip.IP) || bytes.Equal(ipr.ip.IP, ipv4tps.IP_ALL)) && ipr.protocol == proto) {
	//		//Info.Println("Not interested in packet: dropping.")
	//		// TODO should this already have been done in the read manager?
	//		return ipr.ReadFrom()
	//	}

	// verify checksum
	if !verifyIPChecksum(hdr) {
		//Info.Println("Header checksum incorrect, packet dropped")
		return ipr.ReadFrom() // return another packet instead
	}

	packetOffset := uint16(hdr[6]&0x1f)<<8 + uint16(hdr[7])
	//fmt.Println("PACK OFF", packetOffset, "HEADER FLAGS", (hdr[6] >> 5))
	if ((hdr[6]>>5)&0x01 == 0) && (packetOffset == 0) {
		// not a fragment
		return rip, lip, b, p, nil
	} else {
		// is a fragment
		bufID := string([]byte{hdr[12], hdr[13], hdr[14], hdr[15], // the source IP
			hdr[16], hdr[17], hdr[18], hdr[19], // the destination IP
			hdr[9],         // protocol
			hdr[4], hdr[5], // identification
		})
		//Trace.Printf("rcvd a fragment-bufId: %x, len: %d\n", bufID, len(b))

		if _, ok := ipr.fragBuf[bufID]; !ok {
			// create the fragment buffer and quit
			//Trace.Printf("creating a new buffer for %x\n", bufID)
			ipr.fragBuf[bufID] = make(chan []byte, FRAGMENT_ASSEMBLER_BUFFER_SIZE)
			quit := make(chan bool, 1)
			done := make(chan bool, 1)
			didQuit := make(chan bool, 1)

			// create the packet assembler in a goroutine to allow the program to continue
			go ipr.fragmentAssembler(ipr.fragBuf[bufID], quit, didQuit, done)
			go ipr.killFragmentAssembler(quit, didQuit, done, bufID)
		}

		// send the packet to the assembler
		select {
		case ipr.fragBuf[bufID] <- b:
		default:
			logs.Warn.Println("Dropping fragmented packet, no space in fragment buffer")
		}

		// after dealing with the fragment, try reading again
		return ipr.ReadFrom()
	}
}

func (ipr *IP_Reader) Close() error {
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
