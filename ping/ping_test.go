package ping

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
)

func TestPing(t *testing.T) {
	err := GlobalPingManager.SendPing("127.0.0.1", time.Second)
	if err != nil {
		logs.Error.Println(err)
		t.Error(err)
	} else {
		t.Log("Success")
	}
}
