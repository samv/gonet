package main

import (
	"errors"
	"fmt"
	"net"
	"time"
	//    "syscall"
	//"golang.org/x/net/ipv4"
)

type IP_Reader struct {
	incomingPackets chan []byte
	nr              *Network_Reader
	protocol        uint8
	ip              string
	fragBuf         map[string](chan []byte)
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
	}, nil
}

func slicePacket(b []byte) (hrd, payload []byte) {
	hdrLen := int(b[0]&0x0f) * 4
	//fmt.Println("HdrLen: ", hdrLen)
	return b[:hdrLen], b[hdrLen:]
}

const FRAGMENT_TIMEOUT = 15

func (ipr *IP_Reader) ReadFrom() (ip string, b, payload []byte, e error) {
	fmt.Println("STARTING READ")
	b = <-ipr.incomingPackets
	fmt.Println("RAW READ COMPLETED")
	//fmt.Println("Read Length: ", len(b))
	//fmt.Println("Full Read Data: ", b)

	ip = net.IPv4(b[12], b[13], b[14], b[15]).String()
	hdr, p := slicePacket(b)

	packetOffset := uint16(hdr[6]&0x1f)<<8 + uint16(hdr[7])
	fmt.Println("PACK OFF", packetOffset, "HEADER FLAGS", (hdr[6] >> 5))
	if ((hdr[6]>>5)&0x01 == 0) && (packetOffset == 0) { // if not fragment
		// verify checksum
		if calcChecksum(hdr, false) != 0 {
			//fmt.Println("Header checksum verification failed. Packet dropped.")
			//fmt.Println("Wrong header: ", hdr)
			//fmt.Println("Payload (dropped): ", p)
			return "", nil, nil, errors.New("Header checksum incorrect, packet dropped")
		}

		//fmt.Println("Payload Length: ", len(p))
		//fmt.Println("Full payload: ", p)
		fmt.Println("PACKET COMPLETELY READ")
		return ip, b, p, nil
	} else {
		bufID := string([]byte{hdr[12], hdr[13], hdr[14], hdr[15], // the source IP
			hdr[16], hdr[17], hdr[18], hdr[19], // the destination IP
			hdr[9],         // protocol
			hdr[4], hdr[5], // identification
		})

		if c, ok := ipr.fragBuf[bufID]; ok {
			// the fragment has already started
			go func() { c <- b }()
		} else {
			// create the fragment buffer
			ipr.fragBuf[bufID] = make(chan []byte)

			// create the packet assembler in a goroutine to allow the program to continue
			go func(in <-chan []byte, finished chan<- []byte) {
				payload := <-in
				extraFrags := make(map[uint64]([]byte))
				//goalLen := int64(-1)
				t := time.Now()
				for time.Since(t).Seconds() <= FRAGMENT_TIMEOUT {
					select {
					case frag := <-in:
						hdr, p := slicePacket(frag)
						offset := 8 * (uint64(hdr[6]&0x1f)<<8 + uint64(hdr[7]))
						if (hdr[6]>>5)&0x01 == 0 {
							//totalLen := uint64(hdr[2]<<8 + hdr[3])
							//goalLen = int64(offset + totalLen)
						}
						fmt.Println("RECEIVED FRAG")
						fmt.Println("Offset:", offset)
						fmt.Println(len(payload))
						if offset == uint64(len(payload)) {
							payload = append(payload, p...)
							for storedFrag, found := extraFrags[uint64(len(payload))]; found; {
								delete(extraFrags, uint64(len(payload)))
								payload = append(payload, storedFrag...)
							}
							if (hdr[6]>>5)&0x01 == 0 {
								fullPacketHdr := hdr
								totalLen := uint16(fullPacketHdr[0]&0x0F)*4 + uint16(len(payload))
								fullPacketHdr[2] = byte(totalLen >> 8)
								fullPacketHdr[3] = byte(totalLen)
								fullPacketHdr[6] = 0
								fullPacketHdr[7] = 0

								// send the packet back into processing
								go func() {
									finished <- append(fullPacketHdr, payload...)
									fmt.Println("FINISHED")
									fmt.Println(append(fullPacketHdr, payload...))
								}()
								fmt.Println("Just wrote back in")
								return
							}
						} else {
							extraFrags[8*(uint64(p[6])<<3>>11+uint64(p[7]))] = p
						}
					default:
						// make the timeout actually have a chance of being hit
					}
				}

				// drop the packet upon timeout
				fmt.Println(errors.New("Fragments took too long, packet dropped"))
				return
			}(ipr.fragBuf[bufID], ipr.incomingPackets)

			// send in the first fragment
			ipr.fragBuf[bufID] <- p

			// TODO: Remove the fragment buffer after some time
		}

		// after dealing with the fragment, try reading again
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
