package main

import (
	"fmt"
	"network/tcp"
	"network/ipv4"
)

func main() {
	server, err := tcp.NewServer()
	if err != nil {
		fmt.Println("derp", err)
		return
	}
	server.BindListen(80, ipv4.IPAll)
	for {
		socket, _, _, err := server.Accept()
		if err != nil {
			fmt.Println("Dank", err)
			continue
		}
		request, err := socket.Recv(8000)
		if err != nil {
			fmt.Println("swag", err)
			continue
		}
		fmt.Println(request)
		response := "Dank World m8"
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
			fmt.Println("we done kiddos", err)
			continue
		}
	}
}
