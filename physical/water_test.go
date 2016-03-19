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

	const tapName = "tap0"
	ifce, err := water.NewTAP(tapName)
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
			/*logs*/ logs.Info.Println("Write success")
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

func TestReadWater(t *testing.T) {
	const tapName = "tap0"
	ifce, err := water.NewTAP(tapName)
	if err != nil {
		logs.Error.Fatalln(err)
	}

	d := make([]byte, 1522)
	for {
		n, err := ifce.Read(d)
		if err != nil {
			logs.Error.Fatalln(err)
		}
		packet := d[:n]
		logs.Info.Println("Got a packet:", packet)
		if waterutil.IsIPv4(d) {
			logs.Info.Printf("Source:      %v\n", waterutil.IPv4Source(packet))
			logs.Info.Printf("Destination: %v\n", waterutil.IPv4Destination(packet))
			logs.Info.Printf("Protocol:    %v\n", waterutil.IPv4Protocol(packet))
			break
		}
		logs.Info.Print("\n\n")
	}
}
