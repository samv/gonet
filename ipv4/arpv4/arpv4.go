package arpv4

import (
	"errors"
	"network/arp"
	"network/ethernet"
	"network/ipv4/ipv4src"
	"network/ipv4/ipv4tps"

	"sync"

	"github.com/hsheth2/notifiers"
)

type ARPv4_Table struct {
	table         map[ipv4tps.IPhash](*ethernet.MAC_Address)
	tableMutex    *sync.RWMutex
	replyNotifier *notifiers.Notifier
}

func NewARP_Table() (*ARPv4_Table, error) {
	return &ARPv4_Table{
		table:         make(map[ipv4tps.IPhash](*ethernet.MAC_Address)),
		replyNotifier: notifiers.NewNotifier(),
		tableMutex:    &sync.RWMutex{},
	}, nil
}

func (table *ARPv4_Table) Lookup(ip arp.ARP_Protocol_Address) (*ethernet.MAC_Address, error) {
	table.tableMutex.RLock()
	defer table.tableMutex.RUnlock()
	if ans, ok := table.table[ip.(*ipv4tps.IPaddress).Hash()]; ok {
		return ans, nil
	}
	//	d, _ := ip.Marshal()
	//	logs.Error.Printf("ARP lookup into table failed; ip: %v\n", d)
	return nil, errors.New("ARP lookup into table failed")
}

func (table *ARPv4_Table) LookupRequest(ip arp.ARP_Protocol_Address) (*ethernet.MAC_Address, error) {
	x, err := table.Lookup(ip)
	if err == nil {
		return x, nil
	}
	return table.Request(ip)
}

func (table *ARPv4_Table) Request(rip arp.ARP_Protocol_Address) (*ethernet.MAC_Address, error) {
	return arp.GlobalARP_Manager.Request(ethernet.ETHERTYPE_IP, rip)
}

func (table *ARPv4_Table) Add(ip arp.ARP_Protocol_Address, addr *ethernet.MAC_Address) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	d := ip.(*ipv4tps.IPaddress)
	// //ch logs.Trace.Printf("ARPv4 table: add: %v (%v)\n", addr.Data, *d)
	table.tableMutex.Lock()
	table.table[d.Hash()] = addr
	table.tableMutex.Unlock()
	table.GetReplyNotifier().Broadcast(ip)
	return nil
}

func (table *ARPv4_Table) GetReplyNotifier() *notifiers.Notifier {
	return table.replyNotifier
}

func (table *ARPv4_Table) Unmarshal(d []byte) arp.ARP_Protocol_Address {
	return &ipv4tps.IPaddress{IP: d}
}

func (table *ARPv4_Table) GetAddress() arp.ARP_Protocol_Address {
	return ipv4src.External_ip_address
}
