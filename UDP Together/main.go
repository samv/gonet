package udpTogether

import "fmt"

func main() {
    /*nr, err := NewNetwork_Reader();
    if err != nil {
        fmt.Println(err)
        return
    }

    ipr, err := nr.NewIP_Reader("127.0.0.1", 17)
    if err != nil {
        fmt.Println(err)
        return
    }

    for x := 0; x < 1; x++{
        _, p, err := ipr.ReadFrom()
        if err != nil {
            fmt.Println(err)
            continue
        }

        fmt.Println(p)
    }*/

    udp_manager, err := NewUDP_Manager("127.0.0.1")
    if err != nil {
        fmt.Println(err)
        return
    }

    udp, err := udp_manager.NewUDP(20001, 20000)
    if err != nil {
        fmt.Println(err)
        return
    }

    udp.write([]byte{'h', 'i'})

    data, _ := udp.read(13)
    fmt.Println("Reading: ", data)
}
