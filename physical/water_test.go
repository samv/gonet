package physical

import (
	"testing"
	"time"

	"github.com/hsheth2/logs"
	"github.com/hsheth2/water"
	"github.com/hsheth2/water/waterutil"
)

func TestWriteWater(t *testing.T) {
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
			//ch logs.Info.Println("Write success")
			time.Sleep(20 * time.Millisecond)
		}
	}()

	input := make([]byte, 60000)
	n, err := ifce.Read(input)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	//ch logs.Info.Println(input[:n])
}

func TestReadWater(t *testing.T) {
	const TAP_NAME = "tap0"
	ifce, err := water.NewTAP(TAP_NAME)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	d := make([]byte, 1522)
	for {
		_, err := ifce.Read(d)
		if err != nil {
			logs.Error.Fatalln(err)
		}
		//ch logs.Info.Println("Got a packet:", packet)
		if waterutil.IsIPv4(d) {
			//ch logs.Info.Printf("Source:      %v\n", waterutil.IPv4Source(packet))
			//ch logs.Info.Printf("Destination: %v\n", waterutil.IPv4Destination(packet))
			//ch logs.Info.Printf("Protocol:    %v\n", waterutil.IPv4Protocol(packet))
			break
		}
		//ch logs.Info.Println("\n")
	}
}
