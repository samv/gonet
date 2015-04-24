package main

import (
	"fmt"
)

func main() {
	s, err := New_Server_TCB()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.BindListen(20102, "*")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, ip, port, err := s.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connection:", ip, port)

	data, err := conn.Recv(20)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("received data:", data)
}
