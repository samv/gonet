#! /usr/bin/env python

from subprocess import call
import sys

# argv[1] = count, argv[2] = min number of connections, argv[3] = max number of connections argv[4] = interval

if 'tapip' in sys.argv:
	place = 'tapip'
else:
	place = 'golang'

r = range(int(sys.argv[2]), int(sys.argv[3])+1, int(sys.argv[4]))
for i in r:
	call(["./latency_test_concurrent.py", sys.argv[1], str(i), place])
