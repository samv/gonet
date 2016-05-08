package ping

import (
	"bytes"
	"time"

	"github.com/hsheth2/gonet/icmp"

	"github.com/hsheth2/gonet/ipv4"

	"github.com/hsheth2/logs"
	"sync/atomic"
)

const DATA_56_BYTES = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcd"

func (pm *Ping_Manager) ping_response_dealer() {
	for {
		ping := <-pm.reply
		identNum := uint16(ping.Header.Opt >> 16)
		if _, ok := pm.identifiers[identNum]; !ok {
			/*logs*/ logs.Info.Println("Dropped something from response dealer, identnum=", identNum, "options=", ping.Header.Opt)
			continue
		}
		pm.identifiers[identNum] <- ping
	}
}

func sendSinglePing(writer ipv4.Writer, id uint16, seq uint16, timeout time.Duration, reply chan *icmp.Packet) {
	// prepare packet
	packet := &icmp.Header{
		Tp:   icmp.EchoRequest,
		Code: PING_ICMP_CODE,
		Opt:  uint32(id)<<16 | uint32(seq),
		Data: []byte(DATA_56_BYTES), // TODO make legit by putting the timestamp in the data
	}

	// make data
	icmp.SendPacket(writer, packet)

	time1 := time.Now()
	timeoutTimer := time.NewTimer(timeout)
	go func(seqChan chan *icmp.Packet, header *icmp.Header, time1 *time.Time, timer *time.Timer) {
		defer close(seqChan)
		for {
			select {
			case pingResponse := <-seqChan:
				if !bytes.Equal(pingResponse.Header.Data, header.Data) {
					/*logs*/ logs.Info.Println("Dropped packet because header data not equal to ping sent")
					continue
				}
				time2 := time.Now()
				logs.Info.Printf("%d bytes from %v: icmp_seq=%d time=%f ms",
					len(header.Data)+icmp.HeaderMinSize,
					pingResponse.RIP.IP,
					uint16(header.Opt),
					float32(time2.Sub(*time1).Nanoseconds())/1000000) // put ttl
				return
			case <-timer.C:
				logs.Info.Println("Seq num of", uint16(header.Opt), "timed out")
				return
			}
		}
	}(reply, packet, &time1, timeoutTimer)
}

func (pm *Ping_Manager) initIdentifier(terminate chan bool, numPings uint16, pingCount *uint32) (uint16, [](chan *icmp.Packet), error) {
	// get identifier
	pm.currentIdentifier++
	id := pm.currentIdentifier

	// setup sequence number dealer
	pm.identifiers[id] = make(chan *icmp.Packet)
	seqChannel := make([](chan *icmp.Packet), numPings)

	// create go routine function to deal packets
	go sequenceDealer(pm.identifiers[id], seqChannel, terminate, pingCount)

	return id, seqChannel, nil
}

func sequenceDealer(idInput chan *icmp.Packet, seqChan [](chan *icmp.Packet), terminate chan bool, pingCount *uint32) {
	// TODO verify IPs
	for {
		select {
		case <-terminate:
			//			/*logs*/logs.Info.Println("Terminating seq dealer")
			return
		case packet := <-idInput:
			// /*logs*/logs.Info.Println("icmp in =", packet.Header.Opt)
			seqNum := uint32(packet.Header.Opt)
			if seqNum < atomic.LoadUint32(pingCount) - uint32(1) {
				seqChan[seqNum - 1] <- packet
			} else {
				/*logs*/ logs.Info.Println("Dropping bad seq num packet with existing identifier")
			}
		}
	}
}

const FLOOD_INTERVAL = 0 // TODO add a flood option

func (pm *Ping_Manager) SendPing(ip *ipv4.Address, interval, timeout time.Duration, numPings uint16) error {
	terminate := make(chan bool)
	pingCount := uint32(1)
	id, seqChannel, err := pm.initIdentifier(terminate, numPings, &pingCount)
	if err != nil {
		logs.Error.Println(err)
		return err
	}
	defer func() { terminate <- true }()

	// get ip writer
	writer, err := pm.getIP_Writer(ip)
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	for ;pingCount <= uint32(numPings); atomic.AddUint32(&pingCount, 1) {
		seqChannel[pingCount - 1] = make(chan *icmp.Packet)

		sendSinglePing(writer, id, uint16(pingCount), timeout, seqChannel[pingCount - 1]) // function is non-blocking

		// not last
		if uint16(pingCount) != numPings {
			time.Sleep(interval)
		}
	}

	time.Sleep(timeout)
	return nil
}
