#! /bin/bash

make build > /dev/null
go build scaleTest.go
sudo setcap CAP_NET_RAW=epi ./scaleTest
#for i in `seq 1 $1`;
#do
#	#echo "starting $i"
#	(sleep 0.1; ./throughput_client.py > /dev/null 2>&1) &
#done
(time ./scaleTest $1 > throughputTest.out) 2>&1 | ./parseTimeOutput.py $1
