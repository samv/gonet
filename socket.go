package network

import (
	"network/ipv4"
	"sync"
)

type fd int

var fdManager = new(struct {
	current fd
	sync.Mutex
})

func getFd() fd {
	fdManager.Lock()
	defer fdManager.Unlock()
	fdManager.current++
	return fdManager.current
}

type Socket struct {
	fd fd

	Bind   func(port uint16, ip *ipv4.Address) error
}
