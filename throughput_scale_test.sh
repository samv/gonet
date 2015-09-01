#! /bin/bash

make build > /dev/null
go build scaleTest.go
for i in `seq 1 $1`;
do
	#echo "starting $i"
	(./throughput_client.py > /dev/null 2>&1) &
done
sudo setcap CAP_NET_RAW=epi ./scaleTest
(time ./scaleTest $1 > throughputTest.out) 2>&1 | ./parseTimeOutput.py
