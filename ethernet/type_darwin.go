package ethernet

import "syscall"

func getSockAddr(addr *Ethernet_Addr) syscall.Sockaddr {
	return &syscall.SockaddrInet4{}
}
