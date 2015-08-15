// +build !linux

package ethernet

import "syscall"

func getSockAddr(addr *Ethernet_Addr) syscall.Sockaddr {
	panic("getSockAddr not implemented on this platform")
}
