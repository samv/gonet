package tcp

import (
	"fmt"
	"testing"
	"time"
)

const server_port = 20102
const client_port = 20101
const test_ip = "127.0.0.1"

func TestReadWrite(t *testing.T) {
	// TODO make both server and client read and write
	success := make(chan bool, 1)

	data := []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'}

	// server (writes data)
	go func() {
		s, err := New_Server_TCB()
		if err != nil {
			t.Error(err)
			return
		}
		defer s.Close()

		err = s.BindListen(server_port, test_ip)
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
		defer conn.Close()
		fmt.Println("Server Connection:", ip, port)

		fmt.Println("Tester sending data:", data)
		err = conn.Send(data)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// client (reads data)
	go func() {
		client, err := New_TCB_From_Client(client_port, server_port, test_ip)
		if err != nil {
			t.Error("err", err)
			return
		}

		fmt.Println("Client connecting")
		err = client.Connect()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("Client connected")
		defer client.Close()

		time.Sleep(3 * time.Second)
		fmt.Println("Beginning the read")
		out, err := client.Recv(20)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("got data:", out)

		if string(data) == string(out) {
			fmt.Println("Correct output")
			success <- true
		} else {
			t.Error("Wrong output")
		}
	}()

	select {
	case <-success:
		t.Log("Success")
	case <-time.After(10 * time.Second):
		t.Error("Timed out")
	}
	time.Sleep(5 * time.Second) // wait for the goroutines to exit and close the connections
}
