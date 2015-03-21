package main

import (
	"fmt"
	"time"
)

func main() {
	w, err := NewUDP_Writer(20000, 20000, "127.0.0.1")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		err := w.write([]byte{'h', 'e', 'l', 'l', 'o'})
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
