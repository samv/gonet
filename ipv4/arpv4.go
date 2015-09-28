package ipv4

import (
	"errors"
	"network/arp"
	"network/ethernet"

	"sync"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
)

type ARPv4_Table struct {
	table         map[IPhash](*ethernet.MACAddress)
	tableMutex    *sync.RWMutex
	replyNotifier *notifiers.Notifier
}

var GlobalARPv4_Table *ARPv4_Table

func initARPv4Table() *ARPv4_Table {
	// create ARP table
	table, err := NewARP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add loopback ARP entry
	err = table.Add(Loopback_ip_address, ethernet.LoopbackMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add external loopback entry to ARP
	err = table.Add(External_ip_address, ethernet.ExternalMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// register to get packets
	arp.GlobalARP_Manager.Register(ethernet.EtherTypeIP, table)

	return table
}

func NewARP_Table() (*ARPv4_Table, error) {
	return &ARPv4_Table{
		table:         make(map[IPhash](*ethernet.MACAddress)),
		replyNotifier: notifiers.NewNotifier(),
		tableMutex:    &sync.RWMutex{},
	}, nil
}

func (table *ARPv4_Table) Lookup(ip arp.ARP_Protocol_Address) (*ethernet.MACAddress, error) {
	table.tableMutex.RLock()
	defer table.tableMutex.RUnlock()
	if ans, ok := table.table[ip.(*IPAddress).Hash()]; ok {
		return ans, nil
	}
	//	d, _ := ip.Marshal()
	//	logs.Error.Printf("ARP lookup into table failed; ip: %v\n", d)
	return nil, errors.New("ARP lookup into table failed")
}

func (table *ARPv4_Table) LookupRequest(ip arp.ARP_Protocol_Address) (*ethernet.MACAddress, error) {
	x, err := table.Lookup(ip)
	if err == nil {
		return x, nil
	}
	return table.Request(ip)
}

func (table *ARPv4_Table) Request(rip arp.ARP_Protocol_Address) (*ethernet.MACAddress, error) {
	return arp.GlobalARP_Manager.Request(ethernet.EtherTypeIP, rip)
}

func (table *ARPv4_Table) Add(ip arp.ARP_Protocol_Address, addr *ethernet.MACAddress) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	d := ip.(*IPAddress)
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
	return &IPAddress{IP: d}
}

func (table *ARPv4_Table) GetAddress() arp.ARP_Protocol_Address {
	return External_ip_address
}
