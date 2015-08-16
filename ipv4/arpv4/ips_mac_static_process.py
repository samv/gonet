import sys

__author__ = 'hsheth'

def file_get_contents(filename):
    with open(filename) as f:
        return f.read()

for entry in sys.stdin:
    entry = entry.strip().split(' ')
    if 'lo' in entry[1]:
        ifindex = entry[1]
    else:
        ifindex = "tap0"
    print (entry[2].strip() + ' ' +
           file_get_contents("/sys/class/net/" + ifindex + "/ifindex").strip() + ' ' +
           file_get_contents("/sys/class/net/" + entry[1] + "/address").strip())
