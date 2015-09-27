package main

import (
	"network/ipv4/ipv4tps"
	"network/ping"
	"time"

	"github.com/hsheth2/logs"
)

func main() {
	err := ping.GlobalPingManager.SendPing(&ipv4tps.IPAddress{IP: []byte{127, 0, 0, 1}}, ping.FLOOD_INTERVAL, time.Second, 500)
	if err != nil {
		logs.Error.Println(err)
	} else {
		//ch logs.Info.Println("Worked")
	}
	time.Sleep(500 * time.Millisecond)
}
