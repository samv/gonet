#! /bin/bash

sudo pkill runStack
sudo pkill tapip
#go build runStack.go
sudo setcap CAP_NET_RAW=epi ./runStack
./runStack > /dev/null 2>&1 &
sleep 1
sudo ping -W 2 -c $1 -s 1471 -l $2 -i 0.2 -q 10.0.0.3 | tail -n 2 | ./parsePingOutput.py $1 $2
pkill runStack
