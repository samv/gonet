package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func main() {
	//runtime.GOMAXPROCS(strconv.Atoi(os.Args[1]))

	rm, err := NewUDP_Read_Manager()
	if err != nil {
		fmt.Println(err)
		return
	}

	args := os.Args[1:]
	if len(args) > 0 {
		maxNumRoutines := 0
		go func() {
			for {
				maxNumRoutines = max(maxNumRoutines, runtime.NumGoroutine())
				time.Sleep(500 * time.Microsecond)
			}
		}()

		numConns, _ := strconv.Atoi(args[0])
		//fmt.Println("Number of Connections: ", numConns)
		for i := 0; i < numConns; i++ {
			go func(port int) {
				r, err := rm.NewUDP(uint16(port), "127.0.0.1")
				if err != nil {
					fmt.Println(err)
					return
				}
				time.Sleep(3 * time.Second)
				_, err = r.read(MAX_UDP_PACKET_LEN)
				if err != nil {
					fmt.Println(err)
					return
				}
				time.Sleep(time.Second * 8)
			}(20000 + i)
		}

		time.Sleep(3 * time.Second)
		go func() {
			x := exec.Command("python", "multiConnTest.py", "20000", fmt.Sprint(numConns))
			x.Run()
			//fmt.Println("Ran Command")
		}()

		time.Sleep(8 * time.Second)
		fmt.Println(numConns, ":", maxNumRoutines)
	} else {
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
}
