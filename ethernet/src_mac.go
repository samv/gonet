package ethernet

import (
	"github.com/hsheth2/logs"
	"runtime"
	"path"
	"io/ioutil"
	"strings"
	"net"
	"errors"
	"reflect"
)

const ETH_STATIC_MAC_LOAD_FILE = "external_mac.static"

const (
	loopback_internal_index = internal_index(1)
	external_internal_index = internal_index(2)
)

var loopback_mac_address *MAC_Address = &MAC_Address{Data: []byte{0, 0, 0, 0, 0, 0}}
var external_mac_address *MAC_Address

type source_MAC_Table struct {
	table map[internal_index](*MAC_Address)
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
	external_mac_address = &MAC_Address{
		Data: []byte(hw),
	}

	table.add(loopback_internal_index, loopback_mac_address)
	table.add(external_internal_index, external_mac_address)

	return table
}()

func newSource_MAC_Table() (*source_MAC_Table, error) {
	return &source_MAC_Table{
		table: make(map[internal_index](*MAC_Address)),
	}, nil
}

func (smt *source_MAC_Table) search(in internal_index) (*MAC_Address, error) {
	if ans, ok := smt.table[in]; ok {
		return ans, nil
	}
	return nil, errors.New("Failed to find associated MAC address")
}

func (smt *source_MAC_Table) add(in internal_index, mac *MAC_Address) error {
	smt.table[in] = mac // TODO should we prevent overwriting?
	return nil
}

func getInternalIndex(rmac *MAC_Address) internal_index {
	if reflect.DeepEqual(rmac, loopback_mac_address) {
		return loopback_internal_index
	}
	return external_internal_index
}