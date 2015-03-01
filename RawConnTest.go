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
    readUDP(c, 1024)
    writeUDP(c, make([]byte, 0))
    closeUDP(c)
}

type UDP struct {
    open bool
	conn *ipv4.RawConn
}
func openUDP(src, dest int) (*UDP, error) {
    return nil, nil
}
func readUDP(c *UDP, size int) ([]byte, error) {
    return make([]byte, 0), nil
}
func writeUDP(c *UDP, x []byte) error {
    // TODO convert to byte slice: 5 00 00 44 ad 0b 00 00 40 11 72 72 ac 14 02 fd ac 14 00 06
	h := ipv4.ParseHeader()
	c.conn.WriteTo(h, x, nil)
	return nil
}
func closeUDP(c *UDP) error {
    return c.conn.Close()
}