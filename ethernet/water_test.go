package ethernet

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
	"github.com/songgao/water"
)

const TAP_NAME = "tap0"

func TestWater(t *testing.T) {
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
				64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
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
