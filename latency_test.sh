#! /bin/bash

ping -W 2 -c $1 -s 1471 -i 0.3 -q $3 | tail -n 2 | ./parsePingOutput.py $1 $2
