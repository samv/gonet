package ipv4

import (
	"io/ioutil"
	"path"
	"runtime"
	"strings"

	"github.com/hsheth2/logs"
)

const (
	ipv4StaticRouteLoadFile   = "external_ip.static"
	ipv4StaticGatewayLoadFile = "external_gateway.static"
)

const ipv4DefaultNetmask Netmask = 24

// The stack's IP addresses, which are used when sending data
var (
	LoopbackIPAddress *Address
	ExternalIPAddress *Address
)

var (
	externalGateway *Address
)

type routingTable struct {
	// TODO make this thread safe
	table []*Address // ordered by precedence, last one is default
}

func newRoutingTable() (*routingTable, error) {
	return &routingTable{}, nil
}

var globalRoutingTable *routingTable

func initExternalGateway() *Address {
	_, filename, _, _ := runtime.Caller(1)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), ipv4StaticGatewayLoadFile))
	if err != nil {
		logs.Error.Fatalln(err)
	}
	str := strings.TrimSpace(string(data))
	return MakeIP(str)
}

func initLoopbackIP() *Address {
	return MakeIP("127.0.0.1")
}

func initExternalIP() *Address {
	// Load preferences and defaults file
	_, filename, _, _ := runtime.Caller(1)
	data, err := ioutil.ReadFile(path.Join(path.Dir(filename), ipv4StaticRouteLoadFile))
	if err != nil {
		logs.Error.Fatalln(err)
	}
	str := strings.TrimSpace(string(data))
	return MakeIP(str)
}

func initSourceIPTable() *routingTable {
	externalGateway = initExternalGateway()
	LoopbackIPAddress = initLoopbackIP()
	ExternalIPAddress = initExternalIP()

	table, err := newRoutingTable()
	if err != nil {
		logs.Error.Fatalln(err)
	}

	err = table.add(LoopbackIPAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}
	err = table.add(ExternalIPAddress)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	return table
}

func (sipt *routingTable) add(ip *Address) error {
	sipt.table = append(sipt.table, ip) // TODO ensure the entry has not already been inserted
	return nil
}

func ipCompare(baseS, cmpS *Address, netm Netmask) bool {
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

func (sipt *routingTable) Query(dst *Address) (src *Address) {
	if len(sipt.table) == 0 {
		logs.Error.Fatalln("sipt Query: no entries in table")
	}
	//	//ch logs.Trace.Println("Query:", "table:", sipt.table, "len:", len(sipt.table))
	for _, base := range sipt.table {
		//		//ch logs.Trace.Println("Trying query:", base, "compared to", dst)
		if ipCompare(base, dst, ipv4DefaultNetmask) { // TODO determine netmask dynamically
			return base
		}
	}
	return sipt.table[len(sipt.table)-1]
}

func (sipt *routingTable) Gateway(dst *Address) *Address {
	for _, base := range sipt.table {
		if ipCompare(base, dst, ipv4DefaultNetmask) { // TODO determine dynamically
			return dst
		}
	}
	return externalGateway
}
