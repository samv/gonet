package arpv4

import (
	"errors"
	"network/ethernet"
	"network/arp"
)

type ARPv4_Table struct {
	table map[arp.ARP_Protocol_Address](*ethernet.Ethernet_Addr)
}

func NewARP_Table() (*ARPv4_Table, error) {
	return &ARPv4_Table{
		table: make(map[arp.ARP_Protocol_Address](*ethernet.Ethernet_Addr)),
	}, nil
}

func (table *ARPv4_Table) Lookup(ip arp.ARP_Protocol_Address) (*ethernet.Ethernet_Addr, error) {
	if ans, ok := table.table[ip]; ok {
		return ans, nil
	}
	return &ethernet.Ethernet_Addr{}, errors.New("ARP lookup into table failed") // TODO call probe instead
}

func (table *ARPv4_Table) Add(ip arp.ARP_Protocol_Address, addr *ethernet.Ethernet_Addr) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	table.table[ip] = addr
	return nil
}

func (table *ARPv4_Table) Unmarshal([]byte) arp.ARP_Protocol_Address {
	return nil // TODO implement
}
