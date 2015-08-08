package ethernet

import "syscall"

func getSockAddr() syscall.Sockaddr {
	return &syscall.SockaddrInet4{}
}
