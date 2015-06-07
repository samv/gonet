#!/usr/bin/env python

import socket
import sys

print >> sys.stderr , "TEST"

UDP_IP = "127.0.0.1"
UDP_PORT = int(sys.argv[1])

print >> sys.stderr, "Starting"
sock = socket.socket(socket.AF_INET,  # Internet
                     socket.SOCK_DGRAM)  # UDP
print >> sys.stderr, "Made socket"
sock.bind((UDP_IP, UDP_PORT))
print >> sys.stderr, (UDP_IP, UDP_PORT)
sys.stderr.flush()

print >> sys.stderr, "Waiting"
data, addr = sock.recvfrom(5)  # buffer size is 5 bytes
print >> sys.stderr, (data, addr)
print data

sock.close()
