#! /bin/bash

make build
go build throughput_server.go
(sleep 2; ./throughput_client.py) &
time ./throughput_server