package ping

import (
	"bytes"
	"network/icmpp"
	"time"

	"network/ipv4p"

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

func sendSinglePing(writer *ipv4p.IP_Writer, id, seq uint16, timeout time.Duration, reply chan *icmpp.ICMP_In) {
	// prepare packet
	packet := &icmpp.ICMP_Header{
		TypeF: PING_ECHO_REQUEST_TYPE,
		Code:  PING_ICMP_CODE,
		Opt:   uint32(id)<<16 | uint32(seq),
		Data:  []byte("abcdefg"), // TODO make legit by sending 56 bytes of Data and putting the timestamp in the data
	}

	// make data
	data, err := packet.MarshalICMPHeader()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	// send
	err = writer.WriteTo(data)
	if err != nil {
		logs.Error.Println(err)
		return
	}
	time1 := time.Now()
	timeoutTimer := time.NewTimer(timeout)
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
	}(reply, packet, &time1, timeoutTimer)
}

func (pm *Ping_Manager) initIdentifier(terminate chan bool) (id uint16, seqChannel map[uint16](chan *icmpp.ICMP_In), err error) {
	// get identifier
	pm.currentIdentifier++
	id = pm.currentIdentifier

	// setup sequence number dealer
	pm.identifiers[id] = make(chan *icmpp.ICMP_In)
	seqChannel = make(map[uint16](chan *icmpp.ICMP_In))

	// create go routine function to deal packets
	go sequenceDealer(pm.identifiers[id], seqChannel, terminate)

	return id, seqChannel, nil
}

func sequenceDealer(idInput chan *icmpp.ICMP_In, seqChan map[uint16](chan *icmpp.ICMP_In), terminate chan bool) {
	// TODO verify IPs
	for {
		select {
		case <-terminate:
			logs.Info.Println("Terminating seq dealer")
			return
		case packet := <-idInput:
			// logs.Info.Println("icmp in =", packet.Header.Opt)
			seqNum := uint16(packet.Header.Opt)
			if _, ok := seqChan[seqNum]; ok {
				seqChan[seqNum] <- packet
			} else {
				logs.Info.Println("Dropping bad seq num packet with existing identifier")
			}
		}
	}
}

func (pm *Ping_Manager) SendPing(ip string, interval, timeout time.Duration, numPings uint16) error {
	terminate := make(chan bool)
	id, seqChannel, err := pm.initIdentifier(terminate)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// get ip writer
	writer, err := pm.getIP_Writer(ip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	go func() { // TODO needs to be a go routine?
		for i := uint16(1); i <= numPings; i++ {
			seqChannel[i] = make(chan *icmpp.ICMP_In)

			sendSinglePing(writer, id, i, timeout, seqChannel[i]) // function is non-blocking

			// not last
			if i != numPings {
				time.Sleep(interval)
			}
		}
	}()

	time.Sleep(time.Duration(numPings) * timeout)
	terminate <- true
	return nil
}
