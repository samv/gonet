package arpv4

import (
	"network/arp"
	"network/ethernet"

	"github.com/hsheth2/logs"

	//	"bufio"
	//	"net"
	//	"network/ipv4/ipv4tps"
	//	"os"
	//	"path"
	//	"runtime"
	//	"strings"

	"network/ipv4/ipv4src"
)

var GlobalARPv4_Table = func() *ARPv4_Table {
	// create ARP table
	table, err := NewARP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add loopback ARP entry
	err = table.Add(ipv4src.Loopback_ip_address, ethernet.LoopbackMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add external loopback entry to ARP
	err = table.Add(ipv4src.External_ip_address, ethernet.ExternalMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// register to get packets
	arp.GlobalARP_Manager.Register(ethernet.EtherTypeIP, table)

	return table
}()
