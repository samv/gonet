package ping

import (
	"bytes"
	"network/icmpp"
	"time"

	"github.com/hsheth2/logs"
)

func (pm *Ping_Manager) ping_response_dealer() {
	for {
		ping := <-pm.reply
		identNum := uint16(ping.Header.Opt >> 16)
		if _, ok := pm.identifiers[identNum]; !ok {
			logs.Info.Println("Dropped something from response dealer, identnum=", identNum, "options=", ping.Header.Opt)
			continue
		}
		pm.identifiers[identNum] <- ping
	}
}

func (pm *Ping_Manager) SendPing(ip string, interval, timeout time.Duration, numPings uint16) error {
	pm.currentIdentifier++
	ipWriter, err := pm.getIP_Writer(ip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	channel := make(chan *icmpp.ICMP_In)
	pm.identifiers[pm.currentIdentifier] = channel

	seqChannel := make(map[uint16](chan *icmpp.ICMP_In))

	go func() {
		for i := uint16(1); i <= numPings; i++ {
			// prepare packet
			packet := &icmpp.ICMP_Header{
				TypeF: PING_ECHO_REQUEST_TYPE,
				Code:  PING_ICMP_CODE,
				Opt:   uint32(pm.currentIdentifier)<<16 | uint32(i),
				Data:  []byte("abcdefg"), // TODO make legit by sending 56 bytes of Data
			}

			// make data
			data, err := packet.MarshalICMPHeader()
			if err != nil {
				logs.Error.Println(err)
				return
			}

			// send
			err = ipWriter.WriteTo(data)
			if err != nil {
				logs.Error.Println(err)
				return
			}
			time1 := time.Now()
			timeoutTimer := time.NewTimer(timeout)
			seqChannel[i] = make(chan *icmpp.ICMP_In)
			go func(seqChan chan *icmpp.ICMP_In, header *icmpp.ICMP_Header, time1 *time.Time, timer *time.Timer) {
				for {
					select {
					case pingResonse := <-seqChan:
						if !bytes.Equal(pingResonse.Header.Data, header.Data) {
							logs.Info.Println("Dropped packet cuz data !=")
							continue
						}
						time2 := time.Now()
						logs.Info.Printf("%d bytes from %s: icmp_seq=%d time=%f ms",
							len(header.Data)+icmpp.ICMP_Header_MinSize,
							pingResonse.RIP,
							uint16(header.Opt),
							time2.Sub(*time1).Seconds()*1000) // put ttl
						return
					case <-timer.C:
						logs.Info.Println("Seq num of", uint16(header.Opt), "timed out")
						return
					}
				}
			}(seqChannel[i], packet, &time1, timeoutTimer)

			// not last
			if i != numPings {
				time.Sleep(interval)
			}
		}
	}()

	go func(inChan chan *icmpp.ICMP_In, seqChan map[uint16](chan *icmpp.ICMP_In)) {
		// TODO verify IPs
		for {
			icmp_in := <-inChan
			//logs.Info.Println("icmp in =",icmp_in.Header.Opt)
			seqChan[uint16(icmp_in.Header.Opt)] <- icmp_in
		}
		//TODO terminate this go routine
	}(channel, seqChannel)

	time.Sleep(time.Duration(numPings) * timeout)
	return nil
}
