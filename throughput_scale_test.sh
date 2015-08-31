#! /bin/bash

make build
go build scaleTest.go
for i in `seq 1 $1`;
do
	echo "starting $i"
	(sleep 2; ./throughput_client.py) &
done
sudo setcap CAP_NET_RAW=epi ./scaleTest
time ./scaleTest $1
