package main

import "fmt"

func main() {
	man, err := New_TCP_Client_Manager()
	if err != nil {
		fmt.Println("Err:", err)
		return
	}

	client, err := man.New_TCP_Client(20101, 20102, "127.0.0.1")
	if err != nil {
		fmt.Println("err", err)
	} else {
		client.Connect()
	}
}
