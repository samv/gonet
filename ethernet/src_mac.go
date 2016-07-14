package ethernet

import (
	anet "atman/net"
	"errors"
	"fmt"
	"net"

	"github.com/hsheth2/gonet/physical"

	"github.com/hsheth2/logs"

	"bytes"
)

const ethLoadFileStaticMAC = "external_mac.static"

// Static MAC addresses and broadcast addresses
var (
	LoopbackMACAddress       *MACAddress
	ExternalMACAddress       *MACAddress
	LoopbackBroadcastAddress *MACAddress
	ExternalBroadcastAddress *MACAddress
)

type sourceMACTable struct {
	table map[physical.InternalIndex](*MACAddress)
}

var globalSourceMACTable *sourceMACTable

func initSourceTable() *sourceMACTable {
	table, err := newSourceMACTable()
	if err != nil {
		logs.Error.Fatal(err)
	}

	hw, err := net.ParseMAC(string(anet.DefaultDevice.MacAddr))
	if err != nil {
		logs.Error.Fatal(err)
	}

	// init addresses
	LoopbackMACAddress = &MACAddress{Data: []byte{0, 0, 0, 0, 0, 0}}
	ExternalMACAddress = &MACAddress{Data: []byte(hw)}
	fmt.Printf("Our mac address is %s\n", anet.DefaultDevice.MacAddr)
	LoopbackBroadcastAddress = &MACAddress{Data: []byte{0, 0, 0, 0, 0, 0}}
	ExternalBroadcastAddress = &MACAddress{Data: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}

	table.add(physical.LoopbackInternalIndex, LoopbackMACAddress)
	table.add(physical.ExternalInternalIndex, ExternalMACAddress)

	return table
}

func newSourceMACTable() (*sourceMACTable, error) {
	return &sourceMACTable{
		table: make(map[physical.InternalIndex](*MACAddress)),
	}, nil
}

func (smt *sourceMACTable) search(in physical.InternalIndex) (*MACAddress, error) {
	if ans, ok := smt.table[in]; ok {
		return ans, nil
	}
	return nil, errors.New("Failed to find associated MAC address")
}

func (smt *sourceMACTable) add(in physical.InternalIndex, mac *MACAddress) error {
	smt.table[in] = mac
	return nil
}

func getInternalIndex(rmac *MACAddress) physical.InternalIndex {
	if rmac == LoopbackMACAddress || rmac == LoopbackBroadcastAddress {
		return physical.LoopbackInternalIndex
	} else if rmac == ExternalMACAddress || rmac == ExternalBroadcastAddress {
		return physical.ExternalInternalIndex
	} else if bytes.Equal(rmac.Data, LoopbackMACAddress.Data) || bytes.Equal(rmac.Data, LoopbackBroadcastAddress.Data) {
		return physical.LoopbackInternalIndex
	}
	return physical.ExternalInternalIndex
}
