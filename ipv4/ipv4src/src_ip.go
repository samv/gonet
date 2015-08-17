package ipv4src

import (
	"net"
	"path"
	"runtime"
	"strings"
	"io/ioutil"

	"network/ipv4/ipv4tps"

	"github.com/hsheth2/logs"
)

const IPv4_STATIC_IP_LOAD_FILE = "external_ip.static"
const IPv4_DEFAULT_NETMASK = 24

var (
	Loopback_ip_address *ipv4tps.IPaddress = ipv4tps.MakeIP("127.0.0.1")
	External_ip_address *ipv4tps.IPaddress
)

type Source_IP_Table struct {
	// TODO make this thread safe
	table []ipv4tps.IPaddress // ordered by precedence, last one is default
}

func NewSource_IP_Table() (*Source_IP_Table, error) {
	return &Source_IP_Table{}, nil
}

var GlobalSource_IP_Table = func() *Source_IP_Table {
	table, err := NewSource_IP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	// Load preferences and defaults file
	_, filename, _, _ := runtime.Caller(1)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), IPv4_STATIC_IP_LOAD_FILE))
	if err != nil {
		logs.Error.Fatalln(err)
	}
	str := strings.TrimSpace(string(data))
	External_ip_address = ipv4tps.MakeIP(str)

	err = table.add(*Loopback_ip_address)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	err = table.add(*External_ip_address)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	return table
}()

func (sipt *Source_IP_Table) add(ip ipv4tps.IPaddress) error {
	sipt.table = append(sipt.table, ip) // TODO ensure the entry has not already been inserted
	return nil
}

func ipCompare(baseS, cmpS ipv4tps.IPaddress, netm ipv4tps.Netmask) bool {
	base := net.ParseIP(string(baseS))[12:]
	cmp := net.ParseIP(string(cmpS))[12:]

	// TODO take netmask into account
	for i := 0; i < len(base); i++ {
		if base[i] != cmp[i] && base[i] != 0 {
			return false
		}
	}
	return true
}

func (sipt *Source_IP_Table) Query(dst ipv4tps.IPaddress) (src ipv4tps.IPaddress) {
	if len(sipt.table) == 0 {
		logs.Error.Fatalln("sipt Query: no entries in table")
	}
	//	logs.Trace.Println("Query:", "table:", sipt.table, "len:", len(sipt.table))
	for _, base := range sipt.table {
		//		logs.Trace.Println("Trying query:", base, "compared to", dst)
		if ipCompare(base, dst, IPv4_DEFAULT_NETMASK) { // TODO determine netmask dynamically
			return base
		}
	}
	return sipt.table[len(sipt.table)-1]
}
