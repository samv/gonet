package udpp

import (
	"testing"
	"time"
)

const rwport = 20102
const rwIP = "127.0.0.1"

func TestReadWrite(t *testing.T) {
	success := make(chan bool, 1)

	rm, err := NewUDP_Read_Manager()
	if err != nil {
		t.Fatal(err)
	}

	r, err := rm.NewUDP(rwport, rwIP)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data := []byte{'h', 'e', 'l', 'l', 'o'}

	go func() {
		w, err := NewUDP_Writer(20000, rwport, rwIP)
		if err != nil {
			t.Fatal(err)
		}

		err = w.Write(data)
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("Wrote out")
		}

		w.Close()
	}()

	go func() {
		//time.Sleep(10*time.Second)
		p, err := r.Read(MAX_UDP_PACKET_LEN)
		if err != nil {
			t.Error(err)
		}
		t.Log("Output:", string(p))

		if string(p) == string(data) {
			t.Log("Got correct output")
			success <- true
		} else {
			t.Error("Got Wrong Output")
		}
	}()

	select {
	case <-success:
		t.Log("Success")
	case <-time.After(5 * time.Second):
		t.Error("Timed out")
	}
}
