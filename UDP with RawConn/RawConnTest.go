package main

import (
    "fmt"
)

func main() {
    /*fmt.Println(calcChecksum([]byte{
        69,
        0,
        0,
        115,
        0,
        0,
        64,
        0,
        64,
        17,
        184,
        97,
        192,
        168,
        0,
        1,
        192,
        168,
        0,
        199,
    }, true))*/

    manager, _ := NewUDP_Manager("127.0.0.1")
    c := []*UDP{}
    for i := 0; i < 20; i++ {
        udp, _ := manager.NewUDP(((uint16)(20000 + i)), ((uint16)(20000 + i)))
        c = append(c, udp)
        udp.write([]byte{((byte)(100 + i))})
        //        go func() {
        a, _ := udp.read(1)
        fmt.Println("UDP data: ", a)
        //        }()
    }
    for _, element := range c {
        element.close()
    }
}

