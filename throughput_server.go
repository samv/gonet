package main

import (
//	"log"
//	"net/http"
//	_ "net/http/pprof"
	"github.com/hsheth2/logs"
	"network/ipv4/ipv4tps"
	"network/tcp"
	"time"
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
			time.Sleep(5 * time.Second)
			data, err := conn.Recv(100000)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			logs.Info.Println("first 50 bytes of received data:", data[:50])

			conn.Close()
		}()
	}
}

