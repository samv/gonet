package main
import "fmt"

func main() {
    nr, err := NewNetwork_Reader();
    if err != nil {
        fmt.Println(err)
        return
    }

    for x := 0; x < 3; x++{
        buffer := make([]byte, MAX_IP_PACKET_LEN)
        len, err := nr.getNextPacket(buffer)

        if err != nil {
            fmt.Println(err)
            continue
        }

        fmt.Println("Total Length of Packet: ", len)
        fmt.Println(buffer[:len])
        fmt.Println()
    }
}
