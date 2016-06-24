package udp

import (
	"testing"
	"time"

	"github.com/hsheth2/gonet/ipv4"
)

const rwport = 20102

func TestReadWriteLocal(t *testing.T) {
	readWriteTest(t, ipv4.LoopbackIPAddress, 0)
}

func TestReadWriteLocalFragmentation(t *testing.T) {
	readWriteTest(t, ipv4.LoopbackIPAddress, 10)
}

func TestReadWriteExternal(t *testing.T) {
	t.Skip("External tests actually don't work")
	readWriteTest(t, ipv4.ExternalIPAddress, 0)
}

func readWriteTest(t *testing.T, ip *ipv4.Address, exp int) {
	success := make(chan bool, 1)
	wrote := make(chan bool, 1)

	r, err := NewReader(rwport, ip)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data := []byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < exp; i++ {
		data = append(data, data...)
	}

	go func(data []byte) {
		w, err := NewWriter(20000, rwport, ip)
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

		wrote <- true
	}(data)

	go func(data []byte) {
		//time.Sleep(10*time.Second)
		p, err := r.Read(maxUDPPacketLength)
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
	}(data)

	select {
	case <-success:
		<-wrote
		t.Log("Success")
	case <-time.After(5 * time.Second):
		t.Error("Timed out")
	}
}
