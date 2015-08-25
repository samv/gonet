package main

import (
//	"log"
//	"net/http"
//	_ "net/http/pprof"
	"github.com/hsheth2/logs"
	"network/ipv4/ipv4tps"
	"network/tcp"
)

func main() {
//	go func() {
//		log.Println(http.ListenAndServe("localhost:6060", nil))
//	}()

	s, err := tcp.New_Server_TCB()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	err = s.BindListen(49230, ipv4tps.IP_ALL)
	if err != nil {
		logs.Error.Println(err)
		return
	}

	for {
		conn, ip, port, err := s.Accept()
		if err != nil {
			logs.Error.Println(err)
			return
		}
		logs.Info.Println("Connection:", ip, port)

		go func() {
			err = conn.Send([]byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'})
			if err != nil {
				logs.Error.Println(err)
				return
			}

			data, err := conn.Recv(20)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			logs.Info.Println("received data:", data)

			conn.Close()
		}()
	}
	select {}
}

