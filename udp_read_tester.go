package main

import "fmt"
import "time"

func main() {
	rm, err := NewUDP_Read_Manager()
	if err != nil {
		fmt.Println(err)
		return
	}

	r, err := rm.NewUDP(20102, "127.0.0.1")
	if err != nil {
		fmt.Println(err)
		return
	}

	const layout = "2006-01-02 15:04:05.000000"
	for {
		p, err := r.read(MAX_UDP_PACKET_LEN)
		t := time.Now().Format(layout)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("received message:", string(p), " at ", t)
	}
}
