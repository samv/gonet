package ipv4

import (
	"errors"
	"network/arp"
	"network/ethernet"

	"sync"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/notifiers"
)

type arpv4Table struct {
	table         map[Hash](*ethernet.MACAddress)
	tableMutex    *sync.RWMutex
	replyNotifier *notifiers.Notifier
}

var globalARPv4Table *arpv4Table

func initARPv4Table() *arpv4Table {
	// create ARP table
	table, err := newARPv4Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add loopback ARP entry
	err = table.Add(LoopbackIPAddress, ethernet.LoopbackMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// add external loopback entry to ARP
	err = table.Add(ExternalIPAddress, ethernet.ExternalMACAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// register to get packets
	arp.Register(ethernet.EtherTypeIP, table)

	return table
}

func (ip *Address) ARPEqual(other arp.ProtocolAddress) bool {
	return ip.Equal(other.(*Address))
}

func newARPv4Table() (*arpv4Table, error) {
	return &arpv4Table{
		table:         make(map[Hash](*ethernet.MACAddress)),
		replyNotifier: notifiers.NewNotifier(),
		tableMutex:    &sync.RWMutex{},
	}, nil
}

func (table *arpv4Table) Lookup(ip arp.ProtocolAddress) (*ethernet.MACAddress, error) {
	table.tableMutex.RLock()
	defer table.tableMutex.RUnlock()
	if ans, ok := table.table[ip.(*Address).Hash()]; ok {
		return ans, nil
	}
	//	d, _ := ip.Marshal()
	//	logs.Error.Printf("ARP lookup into table failed; ip: %v\n", d)
	return nil, errors.New("ARP lookup into table failed")
}

func (table *arpv4Table) LookupRequest(ip arp.ProtocolAddress) (*ethernet.MACAddress, error) {
	x, err := table.Lookup(ip)
	if err == nil {
		return x, nil
	}
	return table.Request(ip)
}

func (table *arpv4Table) Request(rip arp.ProtocolAddress) (*ethernet.MACAddress, error) {
	return arp.Request(ethernet.EtherTypeIP, rip)
}

func (table *arpv4Table) Add(ip arp.ProtocolAddress, addr *ethernet.MACAddress) error {
	// if _, ok := table.table[ip]; ok {
	// 	return errors.New("Cannot overwrite existing entry")
	// }
	d := ip.(*Address)
	// /*logs*/logs.Trace.Printf("ARPv4 table: add: %v (%v)\n", addr.Data, *d)
	table.tableMutex.Lock()
	table.table[d.Hash()] = addr
	table.tableMutex.Unlock()
	table.GetReplyNotifier().Broadcast(ip)
	return nil
}

func (table *arpv4Table) GetReplyNotifier() *notifiers.Notifier {
	return table.replyNotifier
}

func (table *arpv4Table) Unmarshal(d []byte) arp.ProtocolAddress {
	return &Address{IP: d}
}

func (table *arpv4Table) GetAddress() arp.ProtocolAddress {
	return ExternalIPAddress
}
