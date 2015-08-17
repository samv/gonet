package arpv4

import (
	"errors"
	"network/ethernet"
	"network/arp"
	"network/ipv4/ipv4tps"
	"net"
)

type ARPv4_Table struct {
	table map[ipv4tps.IPaddress](*ethernet.MAC_Address)
}

func NewARP_Table() (*ARPv4_Table, error) {
	return &ARPv4_Table{
		table: make(map[ipv4tps.IPaddress](*ethernet.MAC_Address)),
	}, nil
}

func (table *ARPv4_Table) Lookup(ip arp.ARP_Protocol_Address) (*ethernet.MAC_Address, error) {
	if ans, ok := table.table[*(ip.(*ipv4tps.IPaddress))]; ok {
		return ans, nil
	}
//	d, _ := ip.Marshal()
//	logs.Error.Printf("ARP lookup into table failed; ip: %v\n", d)
	return nil, errors.New("ARP lookup into table failed") // TODO call request instead
}

func (table *ARPv4_Table) Add(ip arp.ARP_Protocol_Address, addr *ethernet.MAC_Address) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	d := ip.(*ipv4tps.IPaddress)
//	logs.Trace.Printf("ARPv4 table: add: %v (%v)\n", addr.Data, *d)
	table.table[*d] = addr
	return nil
}

func (table *ARPv4_Table) Unmarshal(d []byte) arp.ARP_Protocol_Address {
	s := ipv4tps.IPaddress(net.IPv4(d[0], d[1], d[2], d[3]).String())
	return &s
}

func (table *ARPv4_Table) GetAddress() arp.ARP_Protocol_Address {
	addr := ipv4tps.IPaddress("10.0.0.1") // TODO fix this
	return &addr
}
