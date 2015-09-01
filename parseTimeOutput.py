#! /usr/bin/env python

import sys
lines = sys.stdin.readlines()
lines = lines[-3:]

#print lines
def parseLine(s):
    s = s.strip()
    s = s.split("\t")[1]
    s = s[:-1]
    min, sec = s.split("m")
    sec = float(sec)
    sec += int(min) * 60
    return sec

for i, l in enumerate(lines):
    lines[i] = parseLine(l)

print "\t".join(str(l) for l in lines)