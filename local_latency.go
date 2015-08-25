package main

import (
	"network/ping"
	"network/ipv4/ipv4tps"
	"time"
	"github.com/hsheth2/logs"
)

func main() {
	err := ping.GlobalPingManager.SendPing(&ipv4tps.IPaddress{IP: []byte{127,0,0,1}}, ping.FLOOD_INTERVAL, time.Second, 1000)
	if err != nil {
		logs.Error.Println(err)
	} else {
		logs.Info.Println("Worked")
	}
	time.Sleep(500 * time.Millisecond)
}
