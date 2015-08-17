package arpv4

import (
	"network/arp"
	"network/ethernet"

	"github.com/hsheth2/logs"
)

var GlobalARPv4_Table = func() *ARPv4_Table {
	// create ARP table
	table, err := NewARP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// register to get packets
	arp.GlobalARP_Manager.Register(ethernet.ETHERTYPE_IP, table)

	return table
}()
