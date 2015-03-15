package main

import "fmt"

func main() {
	rm, err := NewUDP_Read_Manager("127.0.0.1")
	if err != nil {
		fmt.Println(err)
		return
	}

	r, err := rm.NewUDP(20000)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		p, err := r.read(MAX_UDP_PACKET_LEN)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Payload: ", p)
	}
}
