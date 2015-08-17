package ipv4

import (
	"bufio"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	"network/ipv4/ipv4tps"

	"github.com/hsheth2/logs"
)

const IPv4_IPS_STATIC_FILENAME = "ips.static"
const IPv4_DEFAULT_NETMASK = 24

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
	file, err := os.Open(path.Join(path.Dir(filename), IPv4_IPS_STATIC_FILENAME))
	if err != nil {
		logs.Error.Fatal(err)
	}
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.Split(sc.Text(), " ")

		//		logs.Trace.Println("Source_IP_Table adding", line[2])
		err = table.add(ipv4tps.IPaddress(line[2])) // TODO make sure this array subscript doesn't fail
		if err != nil {
			logs.Error.Fatal(err)
		}
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
