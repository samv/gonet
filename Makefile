# Makefile for go program

all: udp tcp

udp: udp_read udp_write

tcp: tcp_server tcp_client

udp_read:
	go build udp_read_tester.go  udp_reader.go ipv4_reader.go ipv4_common.go network_reader.go myMACAddr.go;

udp_write:
	go build udp_write_tester.go udp_writer.go ipv4_writer.go ipv4_common.go network_writer.go myMACAddr.go;

tcp_server:
	go build tcp_server_tester.go tcp_server.go ipv4_reader.go ipv4_common.go network_reader.go myMACAddr.go;

tcp_client:
	go build tcp_client_tester.go tcp_client.go ipv4_writer.go ipv4_common.go network_writer.go myMACAddr.go;

clean:
	rm ./udp_read_tester
	rm ./udp_write_tester
	rm ./tcp_client_tester
	rm ./tcp_server_tester