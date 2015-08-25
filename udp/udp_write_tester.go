package udp

import (
	"fmt"
	//"time"
	"network/ipv4/ipv4src"
)

func write_tester() {
	w, err := NewUDP_Writer(20000, 20102, ipv4src.Loopback_ip_address)
	if err != nil {
		fmt.Println(err)
		return
	}

	fragmentTest := []byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < 11; i++ {
		fragmentTest = append(fragmentTest, fragmentTest...)
	}
	err = w.Write(fragmentTest)

	/*data := []byte{'h', 'e', 'l', 'l', 'o'}
	    for {
	        err = w.write(data)
	        fmt.Println("Error", err)
	        time.Sleep(500 * time.Millisecond)
	    }
		/*for {
			err := w.write([]byte{'h', 'e', 'l', 'l', 'o'})
			if err != nil {
				fmt.Println(err)
			}

			time.Sleep(500 * time.Millisecond)
		}*/
}
