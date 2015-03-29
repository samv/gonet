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

	/*fragmentTest := []byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < 11; i++ {
		fragmentTest = append(fragmentTest, fragmentTest...)
	}
	err = w.write(fragmentTest)*/

	data := []byte{'h', 'e', 'l', 'l', 'o'}
    for {
        err = w.write(data)
        fmt.Println("Error", err)
        time.Sleep(2000 * time.Millisecond)
    }
	/*for {
		err := w.write([]byte{'h', 'e', 'l', 'l', 'o'})
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(500 * time.Millisecond)
	}*/
}
