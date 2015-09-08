package tcp

import (
	"fmt"
	"testing"
	"time"

	"network/ipv4/ipv4src"
	"network/ipv4/ipv4tps"
)

const server_port = 20102
const client_port = 20101

func TestReadWriteLocal(t *testing.T) {
	read_write_test(t, ipv4tps.MakeIP("127.0.0.1"))
}

func TestReadWriteOverNetwork(t *testing.T) {
	t.Skip("External tests actually don't work")
	read_write_test(t, ipv4src.External_ip_address)
}

func read_write_test(t *testing.T, ip *ipv4tps.IPaddress) {
	// TODO make both server and client read and write
	success := make(chan bool, 1)

	data := []byte{'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!'}

	// server (reads data, initiates close)
	go func() {
		s, err := New_Server_TCB()
		if err != nil {
			t.Error(err)
			return
		}
		defer s.Close()

		err = s.BindListen(server_port, ip)
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
		client, err := New_TCB_From_Client(client_port, server_port, ip)
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

		time.Sleep(1 * time.Second)
		fmt.Println("Client sending data:", data)
		err = client.Send(data)
		if err != nil {
			t.Error(err)
			return
		}

		time.Sleep(1 * time.Second)
		fmt.Println("Client Closing")
		client.Close()
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
