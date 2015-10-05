package ping

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"

	"network/ipv4"
)

func ping_tester(t *testing.T, ip *ipv4.Address, num uint16) {
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
	ping_tester(t, ipv4.MakeIP("127.0.0.1"), 5)
}

func TestTapPing(t *testing.T) {
	ping_tester(t, ipv4.MakeIP("10.0.0.2"), 5)
}

func TestExternalPing(t *testing.T) {
	ping_tester(t, ipv4.MakeIP("192.168.1.2"), 5) // TODO decide dynamically based on ip address
}

func TestPingBing(t *testing.T) {
	t.Skip("Pinging externally does not work yet")
	ping_tester(t, ipv4.MakeIP("204.79.197.200"), 5) // TODO use DNS to determine this IP
}
