package main
import "fmt"

func main() {
    nr, err := NewNetwork_Reader();
    if err != nil {
        fmt.Println(err)
        return
    }

    ipr, err := nr.NewIP_Reader("127.0.0.1", 17)
    if err != nil {
        fmt.Println(err)
        return
    }

    for x := 0; x < 8; x++{
        b, _, err := ipr.ReadFrom()
        if err != nil {
            fmt.Println(err)
            continue
        }

        fmt.Println("Total Length of Packet: ", len(b))
        fmt.Println("Entire packet: ", b)
        fmt.Println()
    }
}
