package main

import (
	//	"log"
	//	"net/http"
	//	_ "net/http/pprof"
	"network/ipv4/ipv4tps"
	"network/tcp"

	"github.com/hsheth2/logs"
)

func main() {
	numConn := 100

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

	count := make(chan bool, numConn)
	done := make(chan bool)

	for i := 0; i < numConn; i++ {
		conn, ip, port, err := s.Accept()
		if err != nil {
			logs.Error.Println(err)
			return
		}
		logs.Info.Println("Connection:", ip, port)

		go func(count chan bool) {
			data, err := conn.Recv(20000)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			logs.Info.Println("first 50 bytes of received data:", data[:50])

			conn.Close()
			logs.Info.Println("connection finished")

			count <- true
			if len(count) >= numConn {
				done <- true
			}
		}(count)
	}
	logs.Info.Println("Exited loop")
	<-done
}
