import sys

__author__ = 'hsheth'

def file_get_contents(filename):
    with open(filename) as f:
        return f.read()

for entry in sys.stdin:
    entry = entry.strip().split(' ')
    print(entry[0], end=' ')
    if 'tap' in entry[1]:
        print(file_get_contents("external_mac.static").strip())
    else:
        print(file_get_contents("/sys/class/net/" + entry[1] + "/address").strip())
