package ethernet

import (
	"bufio"
	"errors"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"bytes"

	"github.com/hsheth2/logs"
)

const IF_INDEX_ERROR = 0
const ETH_STATIC_MAC_LOAD_FILE = "mac.static"

type Source_MAC_Table struct {
	table map[IF_Index](*MAC_Address)
}

var GlobalSource_MAC_Table = func() *Source_MAC_Table {
	table, err := NewSource_MAC_Table()
	if err != nil {
		logs.Error.Fatal(err)
	}

	_, filename, _, _ := runtime.Caller(1)
	file, err := os.Open(path.Join(path.Dir(filename), ETH_STATIC_MAC_LOAD_FILE))
	if err != nil {
		logs.Error.Fatal(err)
	}
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.Split(sc.Text(), " ")

		index, err := strconv.Atoi(line[0])
		if err != nil {
			logs.Error.Fatal()
		}
		if_index := IF_Index(index)

		hw, err := net.ParseMAC(line[1])
		if err != nil {
			logs.Error.Fatal(err)
		}
		mac := &MAC_Address{
			Data: []byte(hw),
		}

		err = table.add(if_index, mac)
		if err != nil {
			logs.Error.Fatal(err)
		}
	}
	return table
}()

func NewSource_MAC_Table() (*Source_MAC_Table, error) {
	return &Source_MAC_Table{
		table: make(map[IF_Index](*MAC_Address)),
	}, nil
}

func (smt *Source_MAC_Table) findByIf(ifindex IF_Index) (*MAC_Address, error) {
	if ans, ok := smt.table[ifindex]; ok {
		return ans, nil
	}
	return nil, errors.New("Failed to find associated MAC address")
}

// will return IF_INDEX_ERROR as IF_Index if find fails
func (smt *Source_MAC_Table) findByMac(mac *MAC_Address) IF_Index {
	for ind, m := range smt.table {
		if bytes.Equal(mac.Data, m.Data) {
			return ind
		}
	}
	return IF_INDEX_ERROR
}

func (smt *Source_MAC_Table) add(ifindex IF_Index, mac *MAC_Address) error {
	smt.table[ifindex] = mac // TODO should we prevent overwriting?
	return nil
}
