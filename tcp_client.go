package main

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
)

func New_TCB_From_Client(local, remote uint16, dstIP string) (*TCB, error) {
	/*write, err := NewIP_Writer(dstIP, TCP_PROTO)
	if err != nil {
		return nil, err
	}*/

	read, err := TCP_Port_Manager.bind(remote, local, dstIP)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p, err := net.ListenPacket(fmt.Sprintf("ip4:%d", TCP_PROTO), dstIP) // only for read, not for write
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r, err := ipv4.NewRawConn(p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("Finished New TCB from Client")
	return New_TCB(local, remote, dstIP, read, r, TCP_CLIENT)
}
