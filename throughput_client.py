#!/usr/bin/env python

import socket
import sys
import time

if 'tapip' in sys.argv:
    TCP_IP = '10.0.0.1'
else:
    TCP_IP = '10.0.0.3'
TCP_PORT = 49230
MESSAGE = "Hello" * 200

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect((TCP_IP, TCP_PORT))
s.send(MESSAGE)
time.sleep(3)
s.close()
print "finished"
