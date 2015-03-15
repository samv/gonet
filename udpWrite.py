import socket
import sys

DST_IP = "127.0.0.1"
if len(sys.argv) <= 1:
    UDP_PORT = 20000
else:
    UDP_PORT = int(sys.argv[1])
MESSAGE = "Hello, World!"

print "UDP target IP:", DST_IP
print "UDP target port:", UDP_PORT
print "message:", MESSAGE

sock = socket.socket(socket.AF_INET,  # Internet
                     socket.SOCK_DGRAM)  # UDP
sock.sendto(MESSAGE, (DST_IP, UDP_PORT))
