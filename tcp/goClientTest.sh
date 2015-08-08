go build tcp_client_tester.go
sudo setcap CAP_NET_RAW=epi tcp_client_tester
./tcp_client_tester
