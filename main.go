package main
import "fmt"

func main() {
    nr, err := NewNetwork_Reader();
    if err != nil {
        fmt.Println(err)
        return
    }

    for {
        buffer := make([]byte, MAX_IP_PACKET_LEN)
        len, err := nr.getNextPacket(buffer)

        if err != nil {
            fmt.Println(err)
            continue
        }

        fmt.Println(buffer[:len])
        fmt.Println()
    }
}
