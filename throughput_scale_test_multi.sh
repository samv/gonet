#! /bin/bash

for i in `seq 1 $1`; do
	./throughput_scale_test.sh $i
done