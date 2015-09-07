#! /bin/bash

rm -f throughputTest.out
for i in `seq $1 $2`; do
	./throughput_scale_test.sh $i
done
