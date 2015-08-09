package ethernet

import "syscall"

func getSockAddr(addr *Ethernet_Addr) syscall.Sockaddr {
	return &syscall.SockaddrLinklayer{
		// Family is automatically set to AF_PACKET
		Protocol: ETHERTYPE_IP,    // should be inherited anyway
		Addr:     addr.MAC.Make(), // sending to myself
		Halen:    ETH_ALEN,        // may not be correct
		Ifindex:  int(addr.IF_index),
	}
}
