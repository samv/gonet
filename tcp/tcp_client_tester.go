package tcp

import (
	"time"

	"github.com/hsheth2/logs"
	"network/ipv4/ipv4tps"
)

func client_tester() {
	client, err := New_TCB_From_Client(20101, 49230, ipv4tps.MakeIP("10.0.0.1"))
	if err != nil {
		logs.Error.Println("err", err)
		return
	}

	err = client.Connect()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	time.Sleep(1 * time.Second)

	err = client.Send([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'})
	if err != nil {
		logs.Error.Println(err)
		return
	}

	logs.Trace.Println("Beginning the read")
	data, err := client.Recv(40)
	if err != nil {
		logs.Error.Println(err)
		return
	}
	logs.Info.Println("got data:", data)

	time.Sleep(10 * time.Millisecond)
	client.Close()
}
