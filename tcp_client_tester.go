package main

import "fmt"

func main() {
	client, err := New_TCP_Client(2000, 2001, "127.0.0.1")
	if err != nil {
		fmt.Println("err", err)
	} else {
		client.Connect()
	}
}
