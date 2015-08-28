#! /usr/bin/env python

from subprocess import call
import sys

for i in range(1, int(sys.argv[2])+1):
	call(["./latency_test.sh", sys.argv[1], str(i)])
