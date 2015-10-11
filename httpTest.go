package main

import (
	"fmt"
	"network/ipv4"
	"network/tcp"
)

func main() {
	server, err := tcp.NewServer()
	if err != nil {
		fmt.Println("tcp server", err)
		return
	}
	server.BindListen(80, ipv4.IPAll)
	for {
		socket, _, _, err := server.Accept()
		if err != nil {
			fmt.Println("tcp accept", err)
			continue
		}
		request, err := socket.Recv(8000)
		if err != nil {
			fmt.Println("socket recv", err)
			continue
		}
		fmt.Println(request)
		response := "Hello World!"
		socket.Send(
			[]byte(
				"HTTP/1.1 200 OK\r\n" +
					"Content-Type: text/plain\r\n" +
					"Content-Length: " + fmt.Sprint(len(response)) + "\r\n" +
					"Connection: close\r\n"),
		)
		socket.Send([]byte("\r\n"))
		socket.Send([]byte(response))
		err = socket.Close()
		if err != nil {
			fmt.Println("socket close", err)
			continue
		}
	}
}
