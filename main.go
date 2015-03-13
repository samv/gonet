package main
import "fmt"

func main() {
    nr, err := NewNetwork_Reader();
    if err != nil {
        fmt.Println(err)
        return
    }

    for {
        buffer := make([]byte, 4096)
        len, err := nr.getNextPacket(buffer)

        if err != nil {
            fmt.Println(err)
        }

        fmt.Println(buffer[:len])
    }
}
