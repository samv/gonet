package arpv4

import (
	"bufio"
	"net"
	"network/ethernet"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/hsheth2/logs"
)

const LOCAL_IPS_AND_MACS_LOAD_FILE = "ips_mac.static"

var GlobalARP_Table = func() *ARP_Table {
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
		ip := string(line[0])

		// parse if index
		index, err := strconv.Atoi(line[1])
		if err != nil {
			logs.Error.Fatal()
		}
		if_index := ethernet.IF_Index(index)

		// parse mac address
		hw, err := net.ParseMAC(line[2])
		if err != nil {
			logs.Error.Fatalln(err)
		}
		mac := &ethernet.MAC_Address{
			Data: []byte(hw),
		}

		// construct mac and if index structure
		enter := &ethernet.Ethernet_Addr{
			MAC:      mac,
			IF_index: if_index,
		}

		// add static ARP entry
		err = table.Static_Add(ip, enter)
		if err != nil {
			logs.Error.Fatalln(err)
		}
	}

	return table
}()
