package main

import (
    "fmt"
)

func main() {
    manager, _ := NewUDP_Manager("127.0.0.1")
    c := []*UDP{}
    for i := 0; i < 20; i++ {
        udp, _ := manager.NewUDP(((uint16)(20000 + i)), ((uint16)(20000 + i)))
        c = append(c, udp)
        udp.write([]byte{((byte)(100 + i))})
//        go func() {
            a, _ := udp.read(1)
            fmt.Println(a)
//        }()
    }
    for _, element := range c {
        element.close()
    }
}

