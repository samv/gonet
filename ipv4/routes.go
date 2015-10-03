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

// RoutingTable stores the different routes that may be used
type RoutingTable struct {
	// TODO make this thread safe
	table []*Address // ordered by precedence, last one is default
}

// GlobalRoutingTable is the routing table that is used by IPv4 and all transport layer protocols
var GlobalRoutingTable *RoutingTable

func newRoutingTable() (*RoutingTable, error) {
	return &RoutingTable{}, nil
}

func initSourceIPTable() *RoutingTable {
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

func (table *RoutingTable) add(ip *Address) error {
	table.table = append(table.table, ip) // TODO ensure the entry has not already been inserted
	return nil
}

// Query returns the source IP address that will be sent from
// given the IP address that should be sent to.
func (table *RoutingTable) Query(dst *Address) (src *Address) {
	if len(table.table) == 0 {
		logs.Error.Fatalln("sipt Query: no entries in table")
	}
	//	//ch logs.Trace.Println("Query:", "table:", sipt.table, "len:", len(sipt.table))
	for _, base := range table.table {
		//		//ch logs.Trace.Println("Trying query:", base, "compared to", dst)
		if ipCompare(base, dst, ipv4DefaultNetmask) { // TODO determine netmask dynamically
			return base
		}
	}
	return table.table[len(table.table)-1]
}

func (table *RoutingTable) gateway(dst *Address) *Address {
	for _, base := range table.table {
		if ipCompare(base, dst, ipv4DefaultNetmask) { // TODO determine dynamically
			return dst
		}
	}
	return externalGateway
}
