package udp

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/hsheth2/gonet/ipv4"
)

const port = 20412

func TestBasic(t *testing.T) {
	t.Skip("This test doesn't work properly")
	//fmt.Println(exec.Command("pwd").Output())

	succeed := make(chan bool, 1)
	go func() {
		output := make(chan []byte, 1)
		go func() {
			cmd := exec.Command("python", "./udp_write_test_helper.py", fmt.Sprint(port))
			out, err := cmd.Output()
			fmt.Println("Cmd finished")
			if err != nil {
				fmt.Println(out)
				t.Fatal(err)
			}
			output <- out
		}()

		fmt.Println("Creating UDP Writer")
		w, err := NewWriter(20000, port, ipv4.LoopbackIPAddress)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()

		// assemble data
		data := []byte{'h', 'e', 'l', 'l', 'o'}
		//		for i := 0; i <= 10; i++ {
		//			data = append(data, data...)
		//		}

		time.Sleep(500 * time.Millisecond)

		// send data
		fmt.Println("Sending")
		_, err = w.Write(data)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Waiting")
		result := <-output
		fmt.Println("Got output")
		if len(result) != len(data) {
			t.Fatal(len(result), "not equal to", len(data))
		} else {
			fmt.Println("Success")
			succeed <- true
		}
	}()

	// crash on timeout
	select {
	case <-succeed:
		fmt.Println("Passed")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout")
	}
}
