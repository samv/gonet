import socket
import sys
import datetime

DST_IP = "127.0.0.1"
MESSAGE = "Hello, World!"

if len(sys.argv) <= 1:
    UDP_DST_PORT = 20000
else:
    UDP_DST_PORT = int(sys.argv[1])

if len(sys.argv) >= 2:
    numConnections = int(sys.argv[2])
else:
    numConnections = 1

s = socket.socket(socket.AF_INET,  # Internet
              socket.SOCK_DGRAM)   # UDP

for x in range(numConnections):
    s.sendto(MESSAGE, (DST_IP, UDP_DST_PORT + x))
