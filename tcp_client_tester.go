package main

import (
	"fmt"
)

func main() {
	client, err := New_TCB_From_Client(20101, 20102, "127.0.0.1")
	if err != nil {
		fmt.Println("err", err)
		return
	}

	err = client.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := client.Recv(20)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("got data:", data)

	err = client.Send([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'})
	if err != nil {
		fmt.Println(err)
		return
	}

	client.Close()
}
