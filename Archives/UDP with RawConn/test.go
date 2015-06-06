package mainTest2

import (
    "syscall"
)

func main() {
    fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
    addr := syscall.SockaddrInet4{
        Port: 0,
        Addr: [4]byte{127, 0, 0, 1},
    }
    syscall.Sendto(fd, []byte{255}, 0, &addr)
}
