import socket
import sys
import datetime

DST_IP = "127.0.0.1"
if len(sys.argv) <= 1:
    UDP_DST_PORT = 20000
else:
    UDP_DST_PORT = int(sys.argv[1])

if len(sys.argv) >= 2:
    MESSAGE = str(sys.argv[2]).zfill(5)
else:
    MESSAGE = "Hello, World!"

#print "UDP target IP:", DST_IP
#print "UDP target port:", UDP_DST_PORT
#print "message:", MESSAGE
print "Sent ", MESSAGE, " at ", datetime.datetime.now()

sock = socket.socket(socket.AF_INET,  # Internet
                     socket.SOCK_DGRAM)  # UDP
sock.sendto(MESSAGE, (DST_IP, UDP_DST_PORT))
