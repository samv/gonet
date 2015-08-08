package arp

import (
	"network/ethernet"
	"github.com/hsheth2/logs"
	"errors"
)

type ARP_Entry struct {
	mac *ethernet.MAC_Address
	ifindex int
}

type ARP_Table struct {
	table map[string](ARP_Entry)
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
		table: make(map[string](ARP_Entry)),
	}, nil
}

func (table *ARP_Table) Lookup(ip string) (ARP_Entry, error) {
	if ans, ok := table.table[ip]; ok {
		return ans, nil
	}
	return ARP_Entry{}, errors.New("ARP lookup into table failed") // TODO call probe instead
}

func (table *ARP_Table) probe(ip string) {
	// TODO implement later
}

func (table *ARP_Table) Static_Add(ip string, mac *ethernet.MAC_Address, ifindex int) error {
	if _, ok := table.table[ip]; ok {
		return errors.New("Cannot overwrite existing entry")
	}
	table.table[ip] = ARP_Entry{mac, ifindex}
	return nil
}
