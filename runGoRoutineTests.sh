#!/usr/bin/env bash

for c in `seq 1 200`; do
	echo "New Number of Connections" $c;
	for t in `seq 1 5`; do
		echo "Trial" $t;
		sudo ./udp_read_tester $c >> "TestResults.txt"
	done
done