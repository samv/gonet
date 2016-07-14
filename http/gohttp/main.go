package main

import (
	"fmt"
	"github.com/hsheth2/gonet/arp"
	"github.com/hsheth2/gonet/http"
	"github.com/hsheth2/gonet/icmp"
	"github.com/hsheth2/gonet/ipv4"
	"runtime/debug"
)

func main() {
	// must run these modules' init() methods
	arp.Noop()
	icmp.Noop()
	ipv4.Noop()
	defer func() {
		if p := recover(); p != nil {
			fmt.Println("**** PANIC ****")
			fmt.Printf("panic was: %s\n", p)
			stack := debug.Stack()
			fmt.Printf("Stack trace: %s\n", stack)
		} else {
			fmt.Println("**** EXITED ****")
		}
	}()
	http.Run()
}
