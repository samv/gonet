#! /bin/bash

make build
go build scaleTest.go
for i in $(eval echo {0..$1});
do
(sleep 2; ./throughput_client.py) &
done
sudo setcap CAP_NET_RAW=epi ./scaleTest
time ./scaleTest $1