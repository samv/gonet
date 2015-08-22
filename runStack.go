package main

import (
	_ "net/http/pprof"
	_ "network/ping"
	"log"
	"net/http"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	select {}
}
