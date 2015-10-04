package tcp

import (
	"fmt"
	"network/ipv4"
	"testing"
	"time"
)

func TestReadWriteLocal(t *testing.T) {
	readWriteTest(t, ipv4.MakeIP("127.0.0.1"))
}

func TestReadWriteOverNetwork(t *testing.T) {
	t.Skip("External tests actually don't work")
	readWriteTest(t, ipv4.ExternalIPAddress)
}

func readWriteTest(t *testing.T, ip *ipv4.Address) {
	const serverPort = 20102
	const clientPort = 20101

	// TODO make both server and client read and write
	success := make(chan bool, 1)

	data := []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'}

	// server (reads data, initiates close)
	go func() {
		s, err := NewServer()
		if err != nil {
			t.Error(err)
			return
		}
		defer s.Close()

		err = s.BindListen(serverPort, ip)
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println("Waiting to accept connection")
		conn, ip, port, err := s.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("Server Connection:", ip, port)

		fmt.Println("Beginning the read")
		out, err := conn.Recv(20)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("got data:", out)

		fmt.Println("Server close")
		conn.Close()
		fmt.Println("Server close finished")

		if string(data) == string(out) {
			fmt.Println("Correct output")
			success <- true
		} else {
			t.Error("Wrong output")
		}
	}()

	// client (sends data)
	go func() {
		client, err := NewClient(clientPort, serverPort, ip)
		if err != nil {
			t.Error("err", err)
			return
		}

		fmt.Println("Client connecting")
		conn, err := client.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("Client connected")

		time.Sleep(1 * time.Second)
		fmt.Println("Client sending data:", data)
		err = conn.Send(data)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(1 * time.Second)
		fmt.Println("Client Closing")
		conn.Close()
		fmt.Println("Client Close finished")
	}()

	select {
	case <-success:
		t.Log("Success")
	case <-time.After(10 * time.Second):
		t.Error("Timed out")
	}
	time.Sleep(5 * time.Second) // wait for the goroutines to exit and close the connections
}
