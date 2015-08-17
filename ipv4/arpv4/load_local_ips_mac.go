package arpv4

import (
	"network/arp"
	"network/ethernet"

	"github.com/hsheth2/logs"
	"runtime"
	"path"
	"os"
	"bufio"
	"strings"
	"network/ipv4/ipv4tps"
	"net"
)

const LOCAL_IPS_AND_MACS_LOAD_FILE = "ips_mac.static"

var GlobalARPv4_Table = func() *ARPv4_Table {
	// create ARP table
	table, err := NewARP_Table()
	if err != nil {
		logs.Error.Fatalln(err)
	}

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
		err = table.Add(&ip, mac)
		if err != nil {
			logs.Error.Fatalln(err)
		}
	}

	// register to get packets
	arp.GlobalARP_Manager.Register(ethernet.ETHERTYPE_IP, table)

	return table
}()
