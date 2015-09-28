package ipv4

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"

	"github.com/hsheth2/logs"
)

const (
	IPv4_STATIC_IP_LOAD_FILE      = "external_ip.static"
	IPv4_STATIC_GATEWAY_LOAD_FILE = "external_gateway.static"
)
const IPv4_DEFAULT_NETMASK = 24

var (
	Loopback_ip_address *IPAddress = MakeIP("127.0.0.1")
	External_ip_address *IPAddress
	external_gateway    *IPAddress = func() *IPAddress {
		_, filename, _, _ := runtime.Caller(1)
		data, err := ioutil.ReadFile(path.Join(path.Dir(filename), IPv4_STATIC_GATEWAY_LOAD_FILE))
		if err != nil {
			logs.Error.Fatalln(err)
		}
		str := strings.TrimSpace(string(data))
		return MakeIP(str)
	}()
)

type Source_IP_Table struct {
	// TODO make this thread safe
	table []*IPAddress // ordered by precedence, last one is default
}

func NewSource_IP_Table() (*Source_IP_Table, error) {
	return &Source_IP_Table{}, nil
}

var GlobalSource_IP_Table *Source_IP_Table

func initSourceIPTable() *Source_IP_Table {
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
	External_ip_address = MakeIP(str)
	// //ch logs.Info.Println("using ext ip:", External_ip_address)

	err = table.add(Loopback_ip_address)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	err = table.add(External_ip_address)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	return table
}

func (sipt *Source_IP_Table) add(ip *IPAddress) error {
	sipt.table = append(sipt.table, ip) // TODO ensure the entry has not already been inserted
	return nil
}

func ipCompare(baseS, cmpS *IPAddress, netm Netmask) bool {
	base := baseS.IP
	cmp := cmpS.IP

	// TODO take netmask into account
	for i := 0; i < len(base); i++ {
		if base[i] != cmp[i] && base[i] != 0 {
			return false
		}
	}
	return true
}

func (sipt *Source_IP_Table) Query(dst *IPAddress) (src *IPAddress) {
	if len(sipt.table) == 0 {
		logs.Error.Fatalln("sipt Query: no entries in table")
	}
	//	//ch logs.Trace.Println("Query:", "table:", sipt.table, "len:", len(sipt.table))
	for _, base := range sipt.table {
		//		//ch logs.Trace.Println("Trying query:", base, "compared to", dst)
		if ipCompare(base, dst, IPv4_DEFAULT_NETMASK) { // TODO determine netmask dynamically
			return base
		}
	}
	return sipt.table[len(sipt.table)-1]
}

func (sipt *Source_IP_Table) Gateway(dst *IPAddress) *IPAddress {
	for _, base := range sipt.table {
		if ipCompare(base, dst, IPv4_DEFAULT_NETMASK) { // TODO determine dynamically
			return dst
		}
	}
	return external_gateway
}
