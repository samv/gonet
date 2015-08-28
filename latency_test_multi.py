#! /usr/bin/env python

from subprocess import call
import sys

# argv[1] = count, argv[2] = max number of connections

if 'tapip' in sys.argv:
	place = 'tapip'
else:
	place = 'golang'

for i in range(2, int(sys.argv[2])+1):
	call(["./latency_test_concurrent.py", sys.argv[1], str(i), place])
