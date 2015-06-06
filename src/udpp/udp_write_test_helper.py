import socket
import sys

print >> sys.stderr, "TEST"

log = open("./udp_test_helper_log.txt", "a")
UDP_IP = "127.0.0.1"
UDP_PORT = int(sys.argv[1])

log.write("Starting\n")
sock = socket.socket(socket.AF_INET,  # Internet
                     socket.SOCK_DGRAM)  # UDP
log.write("Made socket\n")
sock.bind((UDP_IP, UDP_PORT))
print >> log, (UDP_IP, UDP_PORT, "\n")

data, addr = sock.recvfrom(50000)  # buffer size is 50000 bytes
log.write((data, addr))
print data

sock.close()

log.write("\n----------------------------------------------------------------\n")
log.close()