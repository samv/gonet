package main

import (
    "fmt"
//    "net/ipv4"
    "golang.org/x/net/ipv4"
    "os"
)

func main() {
    fmt.Println("Hello, World!")

    c, err := openUDP(5049, 5050)
    if err != nil {
        fmt.Println("Failed")
        os.Exit(1)
    }
    read(c, 1024)
}

//type netConn interface {
//    open(src, dest int) (S UDP, error)
//}

type UDP struct {
	conn *ipv4.RawConn
}
func openUDP(src, dest int) (*UDP, error) {
    return nil, nil
}
func read(S *UDP, size int) ([]byte, error) {
    return make([]byte, 0), nil
}