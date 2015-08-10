package ping

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
	"github.com/pkg/profile"
)

func TestPing(t *testing.T) {
	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	err := GlobalPingManager.SendPing("127.0.0.1", time.Second, time.Second, 5)
	if err != nil {
		logs.Error.Println(err)
		t.Error(err)
	} else {
		t.Log("Success")
	}
	time.Sleep(500 * time.Millisecond)
}
