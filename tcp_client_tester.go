package main

import "time"
import "github.com/hsheth2/logs"
import "network/tcpp"

func main() {
	client, err := tcpp.New_TCB_From_Client(20101, 20199, "127.0.0.1")
	if err != nil {
		logs.Error.Println("err", err)
		return
	}

	err = client.Connect()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	time.Sleep(5 * time.Second)
	logs.Trace.Println("Beginning the read")
	data, err := client.Recv(20)
	if err != nil {
		logs.Error.Println(err)
		return
	}

	logs.Info.Println("got data:", data)

	err = client.Send([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'})
	if err != nil {
		logs.Error.Println(err)
		return
	}

	client.Close()
}
