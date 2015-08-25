#!/usr/bin/env python

import socket
import sys

if 'tapip' in sys.argv:
    TCP_IP = '10.0.0.1'
else:
    TCP_IP = '10.0.0.3'
TCP_PORT = 49230
BUFFER_SIZE = 1024
MESSAGE = "Hello, World!"

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.connect((TCP_IP, TCP_PORT))
s.send(MESSAGE)
data = s.recv(BUFFER_SIZE)
s.close()

print "received data:", data
