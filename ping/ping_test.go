package ping

import (
	"testing"
	"time"

	"network/ipv4"

	"github.com/hsheth2/logs"
)

func ping_tester(t *testing.T, ip ipv4.IPaddress, num uint16) {
	err := GlobalPingManager.SendPing(ip, time.Second, time.Second, num)
	if err != nil {
		logs.Error.Println(err)
		t.Error(err)
	} else {
		t.Log("Success")
	}
	time.Sleep(500 * time.Millisecond)
}

func TestLocalPing(t *testing.T) {
	ping_tester(t, "127.0.0.1", 5)
}

func TestTapPing(t *testing.T) {
	ping_tester(t, "10.0.0.2", 10)
}

func TestExternalPing(t *testing.T) {
	ping_tester(t, "192.168.1.2", 10) // TODO decide dynamically based on ip address
}

