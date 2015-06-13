#!/usr/bin/env bash
rm inputOurs.txt
for i in `seq 1 5000`; do
	python udpWrite.py 20102 $i >> inputOurs.txt;
	echo $i;
done;

sleep 5;
python timeTest.py Ours;

#kill $!; # kill the udp_read_tester
