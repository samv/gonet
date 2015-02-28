package main

import (
    "fmt"
//    "net/ipv4"
    "golang.org/x/net/ipv4"
)

func main() {
    fmt.Println("Hello, World!")
}

type UDP struct {
	conn *ipv4.RawConn
}
