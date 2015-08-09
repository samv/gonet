package arpv4

import (
	"errors"
	"network/ethernet"

	"github.com/hsheth2/logs"
)

type ARP_Table struct {
	table map[string](*ethernet.Ethernet_Addr)
}

var GlobalARP_Table = func() *ARP_Table {
	x, err := NewARP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}
	return x
}()

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
	if _, ok := table.table[ip]; ok {
		return errors.New("Cannot overwrite existing entry")
	}
	table.table[ip] = addr
	return nil
}
