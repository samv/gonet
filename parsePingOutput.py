#! /usr/bin/env python

import sys
lines = sys.stdin.readlines()
dropRate = lines[0]
stats = lines[1]

splitDropRate = dropRate.split(' ')
numSent = splitDropRate[0]
numReceived = splitDropRate[3]

splitStats = stats.split(' ')[3].split('/')
min = splitStats[0]
avg = splitStats[1]
max = splitStats[2]
mdev = splitStats[3]

print sys.argv[1] + "\t" + sys.argv[2] + "\t" +  numSent + "\t" + numReceived + "\t" + min + "\t" + avg + "\t" + max + "\t" + mdev
