package ethernet

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/water"
)

func TestWater(t *testing.T) {
	t.Skip("caution: this test will always fail, as the network_rw will already be using tap0")

	const TAP_NAME = "tap0"
	ifce, err := water.NewTAP(TAP_NAME)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	go func() {
		for i := 0; i < 2000; i++ {
			_, err = ifce.Write([]byte{
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				8, 0,
				69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			})
			if err != nil {
				logs.Error.Fatalln(err)
			}
			logs.Info.Println("Write success")
			time.Sleep(20 * time.Millisecond)
		}
	}()

	input := make([]byte, 60000)
	n, err := ifce.Read(input)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	logs.Info.Println(input[:n])
}
