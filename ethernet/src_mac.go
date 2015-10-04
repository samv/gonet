package ethernet

import (
	"errors"
	"io/ioutil"
	"net"
	"path"
	"runtime"
	"strings"

	"network/physical"

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

	_, filename, _, _ := runtime.Caller(1)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), ethLoadFileStaticMAC))
	if err != nil {
		logs.Error.Fatal(err)
	}
	str := strings.TrimSpace(string(data))
	hw, err := net.ParseMAC(str)
	if err != nil {
		logs.Error.Fatal(err)
	}

	// init addresses
	LoopbackMACAddress = &MACAddress{Data: []byte{0, 0, 0, 0, 0, 0}}
	ExternalMACAddress = &MACAddress{Data: []byte(hw)}
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
