#!/usr/bin/env python

import socket


TCP_IP = '127.0.0.1'
TCP_PORT = 20102
BUFFER_SIZE = 20  # Normally 1024, but we want fast response

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
print "Created socket"

s.bind((TCP_IP, TCP_PORT))
print "Binded"

s.listen(1)
print "Listened"

conn, addr = s.accept()
print "Accepted"

print 'Connection address:', addr
#while 1:
#    data = conn.recv(BUFFER_SIZE)
#    if not data: break
#    print "received data:", data
#    conn.send(data)  # echo

print "Closing"
conn.close()