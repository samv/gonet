package icmp

import (
	"errors"

	"github.com/hsheth2/gonet/ipv4"

	"github.com/hsheth2/logs"
)

var (
	reader  ipv4.Reader
	buffers map[Type](chan *Packet)
)

func init() {
	ipr, err := ipv4.NewReader(ipv4.IPAll, ipv4.IPProtoICMP)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	reader = ipr
	buffers = make(map[Type](chan *Packet))

	go readAll()
}

// Bind allows clients to bind to a specific ICMP type
func Bind(tp Type) (chan *Packet, error) {
	// add the port if not already there
	if _, found := buffers[tp]; !found {
		buffers[tp] = make(chan *Packet, queueSize)
	} else {
		return nil, errors.New("Another application binded")
	}
	return buffers[tp], nil
}

// Unbind allows clients to unbind from ICMP type to stop receiving packets on
// the channel.
func Unbind(tp Type) error {
	// TODO ICMP unbind
	return nil
}

func readAll() {
	for {
		header, err := reader.ReadFrom()
		if err != nil {
			logs.Error.Println(err)
			continue
		}
		///*logs*/logs.Info.Println("Pay", payload, "rip", rip, "lip", lip)

		if len(header.Payload) < HeaderMinSize {
			/*logs*/ logs.Info.Println("Dropping Small ICMP packet:", header.Payload)
			continue
		}

		// extract header
		// TODO verify checksum
		data, err := extractHeader(header.Payload, header.Lip, header.Lip)
		if err != nil {
			/*logs*/ logs.Info.Println(err)
			continue
		}

		if buf, ok := buffers[data.Header.Tp]; ok {
			buf <- data
		} else {
			/*logs*/ logs.Info.Println("Dropping ICMP packet:", header.Payload)
		}
	}
}
