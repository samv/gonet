import sys

__author__ = 'hsheth'

def file_get_contents(filename):
    with open(filename) as f:
        return f.read()

for entry in sys.stdin:
    entry = entry.strip().split(' ')
    print (entry[2].strip() + ' ' + entry[0].strip() + ' ' + file_get_contents("/sys/class/net/" + entry[1] + "/address").strip())