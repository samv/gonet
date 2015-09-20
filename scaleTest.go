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

	"network/ipv4/ipv4src"

	//	"time"
	"github.com/hsheth2/logs"
)

const throughput_port = 49230
const client_port_base = 50000

//const bytes = 1024 // 1 kB
const bytes = 4096 // 4 kB
//const bytes = 20480 // 20 kB
//const bytes = 51200 // 50 kB
//const bytes = 131072 // 128 kB

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

	err = s.BindListenWithQueueSize(throughput_port, ipv4tps.IP_ALL, 10+3*int(numConn))
	if err != nil {
		logs.Error.Println(err)
		return
	}

	count := make(chan bool, numConn+5)
	done := make(chan bool, 2)

	for j := uint16(1); j <= numConn; j++ {
		//ch logs.Info.Println("Connection attempter number", j)
		go func(i uint16) {
			//ch logs.Info.Println("i:", i)
			c, err := tcp.New_TCB_From_Client(client_port_base+i, throughput_port, ipv4src.Loopback_ip_address)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			//ch logs.Info.Println("Client", i, "connecting")
			err = c.Connect()
			if err != nil {
				logs.Error.Println(err)
				return
			}
			//ch logs.Info.Println("Client", i, "connected; proceeding to send data")

			err = c.Send(data)
			if err != nil {
				logs.Error.Println(err)
				return
			}

			//			time.Sleep(50 * time.Millisecond)
			//			//ch logs.Info.Println("Client", i, "starting close")
			//			err = c.Close()
			//			if err != nil {
			//				logs.Error.Println(err)
			//				return
			//			}
			//			//ch logs.Info.Println("Client", i, "finished close")
		}(j)
	}

	//	//ch logs.Info.Println("Signaling done")
	//	done<-true
	//ch logs.Info.Println("About to hit loop")
	for i := uint16(1); i <= numConn; i++ {
		//ch logs.Info.Println("Entering loop")

		//ch logs.Info.Println("Waiting to accept connection")
		conn, ip, port, err := s.Accept()
		if err != nil {
			logs.Error.Println(err, ip, port, i)
			return
		}
		//ch logs.Info.Println("Connection:", ip, port)

		go func(conn *tcp.TCB, count chan bool, num uint16) {
			//ch logs.Info.Println("connection #", num, "attempting to recv")

			revcd := 0
			for {
				data, err := conn.Recv(bytes)
				if err != nil {
					logs.Error.Println(err)
				}

				//ch logs.Info.Println("connection #", num, ": first 30 bytes of received data:", data[:30])
				//ch logs.Info.Println("connection #", num, ": data len =", len(data), "total data len (before this one) = ", revcd)

				revcd += len(data)
				if revcd+2 >= bytes {
					break
				}
			}

			//			err = conn.Close()
			//			if err != nil {
			//				logs.Error.Println(err)
			//			}
			//			//ch logs.Info.Println("connection #", num, "finished")

			count <- true
			if len(count) >= int(numConn) {
				done <- true
			}
			//ch logs.Info.Println("Chan len", len(count))
		}(conn, count, i)
		//ch logs.Info.Println("Loop num", i)
	}
	//ch logs.Info.Println("Exited loop")
	<-done
	//ch logs.Info.Println("Terminating")
}
