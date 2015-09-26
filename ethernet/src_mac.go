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

const ETH_STATIC_MAC_LOAD_FILE = "external_mac.static"

var Loopback_mac_address *MAC_Address = &MAC_Address{Data: []byte{0, 0, 0, 0, 0, 0}}
var External_mac_address *MAC_Address
var Loopback_bcast_address *MAC_Address = &MAC_Address{Data: []byte{0, 0, 0, 0, 0, 0}}
var External_bcast_address *MAC_Address = &MAC_Address{Data: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}

type source_MAC_Table struct {
	table map[physical.InternalIndex](*MAC_Address)
}

var globalSource_MAC_Table = func() *source_MAC_Table {
	table, err := newSource_MAC_Table()
	if err != nil {
		logs.Error.Fatal(err)
	}

	_, filename, _, _ := runtime.Caller(1)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), ETH_STATIC_MAC_LOAD_FILE))
	if err != nil {
		logs.Error.Fatal(err)
	}
	str := strings.TrimSpace(string(data))
	hw, err := net.ParseMAC(str)
	if err != nil {
		logs.Error.Fatal(err)
	}
	External_mac_address = &MAC_Address{
		Data: []byte(hw),
	}

	table.add(physical.LoopbackInternalIndex, Loopback_mac_address)
	table.add(physical.ExternalInternalIndex, External_mac_address)

	return table
}()

func newSource_MAC_Table() (*source_MAC_Table, error) {
	return &source_MAC_Table{
		table: make(map[physical.InternalIndex](*MAC_Address)),
	}, nil
}

func (smt *source_MAC_Table) search(in physical.InternalIndex) (*MAC_Address, error) {
	if ans, ok := smt.table[in]; ok {
		return ans, nil
	}
	return nil, errors.New("Failed to find associated MAC address")
}

func (smt *source_MAC_Table) add(in physical.InternalIndex, mac *MAC_Address) error {
	smt.table[in] = mac // TODO should we prevent overwriting?
	return nil
}

func getInternalIndex(rmac *MAC_Address) physical.InternalIndex {
	if rmac == Loopback_mac_address || rmac == Loopback_bcast_address {
		return physical.LoopbackInternalIndex
	} else if rmac == External_mac_address || rmac == External_bcast_address {
		return physical.ExternalInternalIndex
	} else if bytes.Equal(rmac.Data, Loopback_mac_address.Data) || bytes.Equal(rmac.Data, Loopback_bcast_address.Data) {
		return physical.LoopbackInternalIndex
	}
	return physical.ExternalInternalIndex
}
