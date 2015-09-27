package ipv4

import (
	"time"

	"github.com/hsheth2/logs"
)

func (ipr *ipReader) fragAssembler(in <-chan []byte, quit <-chan bool, didQuit chan<- bool, done chan bool) {
	payload := []byte{}
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

func (ipr *ipReader) killFragAssembler(quit chan<- bool, didQuit <-chan bool, done <-chan bool, bufID string) {
	// sends quit to the assembler if it doesn't send done
	select {
	case <-time.After(fragmentationTimeout):
		//Trace.Println("Force quitting packet assembler")
		quit <- true
		<-didQuit // will block until it has been received
	case <-done:
		//Trace.Println("Received done msg.")
	}

	//Trace.Println("Frag Assemble Ended, finished")
	ipr.fragBufMutex.Lock()
	defer ipr.fragBufMutex.Unlock()
	delete(ipr.fragBuf, bufID)
}
