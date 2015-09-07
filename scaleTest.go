package main

import (
	//	"log"
	//	"net/http"
	//	_ "net/http/pprof"
	"network/ipv4/ipv4tps"
	"network/tcp"

	"os"
	"strconv"
	//"time"

	"github.com/hsheth2/logs"
	"network/ipv4/ipv4src"
)

const throughput_port = 49230
const client_port_base = 50000
const bytes = 1048576 // 1 MB

func main() {
	numc, _ := strconv.Atoi(os.Args[1])
	numConn := uint16(numc)
	data := make([]byte, bytes)

	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()

	s, err := tcp.New_Server_TCB()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	err = s.BindListenWithQueueSize(throughput_port, ipv4tps.IP_ALL, 10+int(numConn))
	if err != nil {
		logs.Error.Println(err)
		return
	}

	count := make(chan bool, numConn)
	done := make(chan bool, 2)

	for i := uint16(1); i <= numConn; i++ {
		logs.Trace.Println("Connection attempter number", i)
		go func(){
			c, err := tcp.New_TCB_From_Client(throughput_port, client_port_base + i, ipv4src.Loopback_ip_address)
			if err != nil {
				logs.Error.Println(err)
			}
			err = c.Send(data)
			if err != nil {
				logs.Error.Println(err)
			}

			err = c.Close()
			if err != nil {
				logs.Error.Println(err)
			}
		}()
	}

//	logs.Trace.Println("Signaling done")
//	done<-true
	logs.Trace.Println("About to hit loop")
	for i := uint16(1); i <= numConn; i++ {
		logs.Trace.Println("Entering loop")
		conn, _, _, err := s.Accept()
		if err != nil {
			logs.Error.Println(err)
			return
		}
		//logs.Info.Println("Connection:", ip, port)

		go func(conn *tcp.TCB, count chan bool) {
			data, err := conn.Recv(bytes)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			logs.Info.Println("first 50 bytes of received data:", data[:50])

			//time.Sleep(500 * time.Millisecond)
			conn.Close()
			logs.Trace.Println("connection finished")

			count <- true
			if len(count) >= int(numConn) {
				done <- true
			}
			logs.Info.Println("Chan len", len(count))
		}(conn, count)
		logs.Info.Println("Loop num", i)
	}
	logs.Info.Println("Exited loop")
	<-done
	logs.Trace.Println("Terminating")
}
