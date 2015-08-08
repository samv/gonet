package ethernet

import "syscall"

func getSockAddr() syscall.Sockaddr {
	return &syscall.SockaddrLinklayer{
		// Family is automatically set to AF_PACKET
		Protocol: ETHERTYPE_IP, // should be inherited anyway
		Addr:     myMACAddr,    // sending to myself
		Halen:    ETH_ALEN,     // may not be correct
		Ifindex:  MyIfIndex,    // TODO: don't hard code this... fix it later
	}
}
