package icmp

import (
	"errors"
	"network/ipv4"

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

func Bind(tp Type) (chan *Packet, error) {
	// add the port if not already there
	if _, found := buffers[tp]; !found {
		buffers[tp] = make(chan *Packet, queueSize)
	} else {
		return nil, errors.New("Another application binded")
	}
	return buffers[tp], nil
}

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
		////ch logs.Info.Println("Pay", payload, "rip", rip, "lip", lip)

		if len(header.Payload) < HeaderMinSize {
			//ch logs.Info.Println("Dropping Small ICMP packet:", payload)
			continue
		}

		// extract header
		// TODO verify checksum
		data, err := ExtractICMPHeader(header.Payload, header.Lip, header.Lip)
		if err != nil {
			//ch logs.Info.Println(err)
			continue
		}

		if buf, ok := buffers[data.Header.Tp]; ok {
			buf <- data
		} else {
			//ch logs.Info.Println("Dropping ICMP packet:", payload)
		}
	}
}
