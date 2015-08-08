package ping

import (
	"bytes"
	"github.com/hsheth2/logs"
	"network/icmpp"
	"network/ipv4p"
	"time"
)

const (
	PING_ECHO_REQUEST_TYPE = 8
	PING_ECHO_REPLY_TYPE   = 0
	PING_ICMP_CODE         = 0
	PING_START_ID          = 8000
)

type Ping_Manager struct {
	// Responding to pings
	input  chan *icmpp.ICMP_In
	output map[string](*ipv4p.IP_Writer)

	// Sending pings
	reply             chan *icmpp.ICMP_In
	currentIdentifier uint16
	identifiers       map[uint16](chan *icmpp.ICMP_In)
}

func NewPing_Manager(icmprm *icmpp.ICMP_Read_Manager) (*Ping_Manager, error) {
	input, err := icmprm.Bind(8)
	if err != nil {
		return nil, err
	}

	reply, err := icmprm.Bind(0)
	if err != nil {
		return nil, err
	}

	pm := &Ping_Manager{
		input:             input,
		output:            make(map[string](*ipv4p.IP_Writer)),
		reply:             reply,
		currentIdentifier: PING_START_ID,
		identifiers:       make(map[uint16](chan *icmpp.ICMP_In)),
	}

	go pm.ping_replier()
	go pm.ping_response_dealer()

	return pm, nil
}

var GlobalPingManager = func() *Ping_Manager {
	pm, err := NewPing_Manager(icmpp.GlobalICMPReadManager)
	if err != nil {
		logs.Error.Fatal(err)
	}
	return pm
}()

func (pm *Ping_Manager) getIP_Writer(ip string) (*ipv4p.IP_Writer, error) {
	if _, ok := pm.output[ip]; !ok {
		wt, err := ipv4p.NewIP_Writer(ip, ipv4p.ICMP_PROTO)
		if err != nil {
			return nil, err
		}
		pm.output[ip] = wt
	}
	return pm.output[ip], nil
}

func (pm *Ping_Manager) ping_response_dealer() {
	for {
		ping := <-pm.reply
		// TODO verify checksum as well
		identNum := uint16(ping.Header.Opt >> 16)
		if _, ok := pm.identifiers[identNum]; !ok {
			logs.Info.Println("Dropped something from response dealer, identnum=", identNum, "options=", ping.Header.Opt)
			continue
		}
		pm.identifiers[identNum] <- ping
	}
}

func (pm *Ping_Manager) ping_replier() {
	for {
		ping := <-pm.input
		wr, err := pm.getIP_Writer(ping.RIP)
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		//logs.Info.Println("replying:", ping)
		go pm.respondTo(wr, ping)
	}
}
func (pm *Ping_Manager) respondTo(writer *ipv4p.IP_Writer, ping *icmpp.ICMP_In) error {
	header := ping.Header
	header.TypeF = PING_ECHO_REPLY_TYPE

	// make packet
	bts, err := header.MarshalICMPHeader()
	if err != nil {
		logs.Error.Println(err)
		return err
	}

	// send
	err = writer.WriteTo(bts)
	if err != nil {
		logs.Error.Println(err)
	}

	return nil
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
				Data:  []byte("abcdefg"), // TODO legitify
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
		//	TODO term plz
	}(channel, seqChannel)

	time.Sleep(time.Duration(numPings) * timeout)
	return nil
}
