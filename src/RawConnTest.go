package main

import (
    "fmt"
    "net/ipv4"
)

func main() {
    fmt.Println("Hello, World!")
}

type UDP struct {
	conn *RawConn
}