package udp

import (
	"network/ipv4"
	"testing"
	"time"
)

const rwport = 20102

func TestReadWriteLocal(t *testing.T) {
	read_write_test(t, ipv4.LoopbackIPAddress, 0)
}

func TestReadWriteLocalFragmentation(t *testing.T) {
	read_write_test(t, ipv4.LoopbackIPAddress, 10)
}

func TestReadWriteExternal(t *testing.T) {
	t.Skip("External tests actually don't work")
	read_write_test(t, ipv4.ExternalIPAddress, 0)
}

func read_write_test(t *testing.T, ip *ipv4.Address, exp int) {
	success := make(chan bool, 1)

	r, err := NewUDP(GlobalUDP_Read_Manager, rwport, ip)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data := []byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < exp; i++ {
		data = append(data, data...)
	}

	go func() {
		w, err := NewUDP_Writer(20000, rwport, ip)
		if err != nil {
			t.Fatal(err)
		}

		_, err = w.Write(data)
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Wrote the data:", data)
		}

		w.Close()
	}()

	go func() {
		//time.Sleep(10*time.Second)
		p, err := r.Read(MAX_UDP_PACKET_LEN)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("Output:", string(p))

		if string(p) == string(data) {
			t.Log("Got correct output:", p)
			success <- true
		} else {
			t.Error("Got Wrong Output:", p)
		}
	}()

	select {
	case <-success:
		t.Log("Success")
	case <-time.After(5 * time.Second):
		t.Error("Timed out")
	}
}
