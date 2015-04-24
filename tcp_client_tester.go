package main

import "fmt"

func main() {
	client, err := New_TCB_From_Client(20101, 20102, "127.0.0.1")
	if err != nil {
		fmt.Println("err", err)
		return
	}

	client.Connect()
}
