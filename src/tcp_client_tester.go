package main

import "time"

func main() {
	client, err := New_TCB_From_Client(20101, 20102, "127.0.0.1")
	if err != nil {
		Error.Println("err", err)
		return
	}

	err = client.Connect()
	if err != nil {
		Error.Println(err)
		return
	}

	time.Sleep(5*time.Second)
	Trace.Println("Beginning the read")
	data, err := client.Recv(20)
	if err != nil {
		Error.Println(err)
		return
	}

	Info.Println("got data:", data)

	err = client.Send([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'})
	if err != nil {
		Error.Println(err)
		return
	}

	client.Close()
}
