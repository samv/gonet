#! /bin/bash

sudo ping -W 2 -c $1 -s 1471 -i 0.2 -q 10.0.0.3 | tail -n 2 | ./parsePingOutput.py $1 $2
