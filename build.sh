#! /bin/bash

#pwd = current directory

go build udp_read_tester.go udp_reader.go ipv4_reader.go ipv4_common.go network_reader.go;
go build udp_write_tester.go udp_writer.go ipv4_writer.go ipv4_common.go;
