package tcpp

import (
	"github.com/hsheth2/logs"
)

func server_tester() {
	s, err := New_Server_TCB()
	if err != nil {
		logs.Error.Println(err)
		return
	}

	err = s.BindListen(20102, "*")
	if err != nil {
		logs.Error.Println(err)
		return
	}

	conn, ip, port, err := s.Accept()
	if err != nil {
		logs.Error.Println(err)
		return
	}
	logs.Info.Println("Connection:", ip, port)

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
}
