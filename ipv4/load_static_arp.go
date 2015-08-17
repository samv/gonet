package ipv4

import (
	"bufio"
	"net"
	"network/ethernet"
	"os"
	"path"
	"runtime"
	"strings"

	"network/ipv4/arpv4"

	"github.com/hsheth2/logs"
	"network/ipv4/ipv4tps"
)

const LOCAL_IPS_AND_MACS_LOAD_FILE = "arpv4/ips_mac.static"

var loaded_static_arp = func() bool { // TODO find better way to do this
	// open file
	_, filename, _, _ := runtime.Caller(1)
	file, err := os.Open(path.Join(path.Dir(filename), LOCAL_IPS_AND_MACS_LOAD_FILE))
	if err != nil {
		logs.Error.Fatalln(err)
	}
	sc := bufio.NewScanner(file)

	for sc.Scan() {
		line := strings.Split(sc.Text(), " ")

		// parse ip address
		ip := ipv4tps.IPaddress(line[0])

		// parse mac address
		hw, err := net.ParseMAC(line[1])
		if err != nil {
			logs.Error.Fatalln(err)
		}
		mac := &ethernet.MAC_Address{
			Data: []byte(hw),
		}

		// add static ARP entry
		err = arpv4.GlobalARPv4_Table.Add(&ip, mac)
		if err != nil {
			logs.Error.Fatalln(err)
		}
	}

	return true
}()
