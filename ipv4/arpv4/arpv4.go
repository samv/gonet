package arpv4

import (
	"errors"
	"network/ethernet"
)

type ARP_Table struct {
	table map[string](*ethernet.Ethernet_Addr)
}

func NewARP_Table() (*ARP_Table, error) {
	return &ARP_Table{
		table: make(map[string](*ethernet.Ethernet_Addr)),
	}, nil
}

func (table *ARP_Table) Lookup(ip string) (*ethernet.Ethernet_Addr, error) {
	if ans, ok := table.table[ip]; ok {
		return ans, nil
	}
	return &ethernet.Ethernet_Addr{}, errors.New("ARP lookup into table failed") // TODO call probe instead
}

func (table *ARP_Table) probe(ip string) {
	// TODO implement later
}

func (table *ARP_Table) Static_Add(ip string, addr *ethernet.Ethernet_Addr) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	table.table[ip] = addr
	return nil
}
