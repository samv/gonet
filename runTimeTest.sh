#!/usr/bin/env bash

./udp_read_tester >> outputOurs.txt & : ;

for i in `seq 1 10`; do
	python udpWrite.py 2010 $i >> inputOurs.txt;
	echo $i;
done;

sleep 5;
python timeTest.py Ours;

kill %%; # kill the udp_read_tester
