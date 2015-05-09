package main

import (
	"net"
	"time"
	//"errors"
	//"fmt"
	//"syscall"
	//"golang.org/x/net/ipv4"
)

type IP_Reader struct {
	incomingPackets chan []byte
	nr              *Network_Reader
	protocol        uint8
	ip              string
	fragBuf         map[string](chan []byte)
	//fragQuit        map[string](chan bool)
}

func (nr *Network_Reader) NewIP_Reader(ip string, protocol uint8) (*IP_Reader, error) {
	c, err := nr.bind(ip, protocol)
	if err != nil {
		return nil, err
	}

	return &IP_Reader{
		incomingPackets: c,
		nr:              nr,
		protocol:        protocol,
		ip:              ip,
		fragBuf:         make(map[string](chan []byte)),
		//fragQuit:        make(map[string](chan bool)),
	}, nil
}

func slicePacket(b []byte) (hrd, payload []byte) {
	hdrLen := int(b[0]&0x0f) * 4
	//fmt.Println("HdrLen: ", hdrLen)
	return b[:hdrLen], b[hdrLen:]
}

func fragmentAssembler(in <-chan []byte, quit <-chan bool, didQuit chan<- bool, finished chan<- []byte, done chan bool) {
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
				check := calculateChecksum(fullPacketHdr[:20])
				fullPacketHdr[10] = byte(check >> 8)
				fullPacketHdr[11] = byte(check)

				// send the packet back into processing
				go func() {
					finished <- append(fullPacketHdr, payload...)
					//fmt.Println("FINISHED")
				}()
				//Trace.Println("Just wrote back in")
				done <- true
				return // from goroutine
			}
			//Trace.Println("Looping")
		}
	}

	// drop the packet upon timeout
	//Trace.Println(errors.New("Fragments took too long, packet dropped"))
	return
}

func killFragmentAssembler(quit chan<- bool, didQuit <-chan bool, done <-chan bool, bufID string) {
	// sends quit to the assembler if it doesn't send done
	select {
	case <-time.After(time.Second * FRAGMENT_TIMEOUT):
		//Trace.Println("Force quitting packet assembler")
		quit <- true
		<-didQuit // will block until it has been received
	case <-done:
		//Trace.Println("Recieved done msg.")
	}

	//Trace.Println("Frag Assemble Ended, finished")
	// TODO: clean the buffer for bufID
}

func (ipr *IP_Reader) ReadFrom() (rip, lip string, b, payload []byte, e error) {
	//fmt.Println("STARTING READ")
	b = <-ipr.incomingPackets
	//fmt.Println("RAW READ COMPLETED")
	//fmt.Println("Read Length: ", len(b))
	//fmt.Print(".")
	//fmt.Println("Full Read Data: ", b)

	hdr, p := slicePacket(b)

	// extract source IP and protocol
	rip = net.IPv4(hdr[12], hdr[13], hdr[14], hdr[15]).String()
	lip = net.IPv4(hdr[16], hdr[17], hdr[18], hdr[19]).String()
	proto := uint8(hdr[9])
	if !((ipr.ip == rip || ipr.ip == "*") && ipr.protocol == proto) {
		//Info.Println("Not interested in packet: dropping.")
		return ipr.ReadFrom()
	}

	// verify checksum
	if verifyChecksum(hdr) != 0 {
		//Info.Println("Header checksum incorrect, packet dropped")
		return ipr.ReadFrom() // return another packet instead
	}

	packetOffset := uint16(hdr[6]&0x1f)<<8 + uint16(hdr[7])
	//fmt.Println("PACK OFF", packetOffset, "HEADER FLAGS", (hdr[6] >> 5))
	if ((hdr[6]>>5)&0x01 == 0) && (packetOffset == 0) {
		// not a fragment
		//fmt.Println("Payload Length: ", len(p))
		//fmt.Println("Full payload: ", p)
		//fmt.Println("PACKET COMPLETELY READ")
		return rip, lip, b, p, nil
	} else {
		// is a fragment
		bufID := string([]byte{hdr[12], hdr[13], hdr[14], hdr[15], // the source IP
			hdr[16], hdr[17], hdr[18], hdr[19], // the destination IP
			hdr[9],         // protocol
			hdr[4], hdr[5], // identification
		})
		//Trace.Printf("rcv a fragment-bufId: %x, len: %d\n", bufID, len(b))

		if c, ok := ipr.fragBuf[bufID]; ok {
			// the fragment has already started
			//Trace.Printf("sending to already existing assembler %x\n", bufID)
			go func() { c <- b }()
		} else {
			// create the fragment buffer and quit
			//Trace.Printf("creating a new buffer for %x\n", bufID)
			ipr.fragBuf[bufID] = make(chan []byte)
			quit := make(chan bool)
			done := make(chan bool)
			didQuit := make(chan bool)

			// create the packet assembler in a goroutine to allow the program to continue
			go fragmentAssembler(ipr.fragBuf[bufID], quit, didQuit, ipr.incomingPackets, done)
			go killFragmentAssembler(quit, didQuit, done, bufID)

			// send in the first fragment
			ipr.fragBuf[bufID] <- b
		}

		// after dealing with the fragment, try reading again
		//fmt.Println("RECURSE")
		return ipr.ReadFrom()
	}
}

func (ipr *IP_Reader) Close() error {
	return ipr.nr.unbind(ipr.ip, ipr.protocol)
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
