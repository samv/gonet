package ping

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
)

func TestPing(t *testing.T) {
	err := GlobalPingManager.SendPing("127.0.0.1", time.Second, time.Second, 5)
	if err != nil {
		logs.Error.Println(err)
		t.Error(err)
	} else {
		t.Log("Success")
	}
	time.Sleep(500 * time.Millisecond)
}
