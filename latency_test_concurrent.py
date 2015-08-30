#! /usr/bin/env python

import subprocess
import sys
import time

# argv[1] = count, argv[2] = concurrent

cmd = "./latency_test_stack_init.sh"
ip = "10.0.0.3"
if 'tapip' in sys.argv:
	cmd = './latency_tapip_stack_init.sh'
	ip = "10.0.0.1"
stack = subprocess.Popen(cmd)
time.sleep(1)

count = int(sys.argv[1])
concurrent = int(sys.argv[2])
r = range(0, concurrent)

ps = []
for i in r:
	ps.append(subprocess.Popen(["./latency_test.sh", str(count), str(concurrent), ip], stdout=subprocess.PIPE))

for i in r:
	ps[i].wait()

line = []
for i in r:
	line.append(ps[i].stdout.readlines()[0].strip().split("\t"))
col = zip(*line)
new = [sum(float(z) for z in x) for x in col]
new[0] = int(new[0])
new[1] = int(col[1][0])
new[2] = int(new[2])
new[3] = int(new[3])
new[4] = min(float(y) for y in col[4])
new[5] = (sum(float(y) for y in col[5]) / len(col[5])) # average
new[6] = max(float(y) for y in col[6])
new = new[:-1] # remove mdev

print '\t'.join(str(w) for w in new)

try:
	stack.kill()
	stack.terminate()
except:
	pass