import datetime
import dateutil.parser as parse
import sys

def trunc(f, nl):
    with open(f, "r") as i:
        data = [x.strip() for x in i.readlines()]
    with open(f, "w") as i:
        i.write("\n".join(x.strip() for x in data[:nl]))
        i.write("\n")

if len(sys.argv) > 1:
    arg = str(sys.argv[1]).strip()
else:
    arg = "Ours"

f1 = "./output{name}.txt".format(name=arg)
f2 =  "./input{name}.txt".format(name=arg)

if len(sys.argv) > 2:
    l = int(sys.argv[2])
    trunc(f1, l)
    trunc(f2, l)
    print "Truncated to new len of", l
    sys.exit(0)

with open(f1) as outputF:
    with open(f2) as inputF:
        out = outputF.readlines()
        inp = inputF.readlines()

#out = [l.strip().split("  ")[2].strip() for l in out]
outN = [l.strip().split("  ") for l in out]

out = {}
for dat in outN:
    dat[0] = int(dat[0].split(" ")[2])
    out[dat[0]] = parse.parse(dat[2])

inpN = [l.strip().split("  ") for l in inp]
inp = []
for dat in inpN:
    inp.append((int(dat[1].strip()), parse.parse(dat[3]) ))

dropped = 0
num = 0
diff = datetime.timedelta(0)
for id, tmIn in inp:
    #print "Processing", id
    if id in out:
        num = num + 1
        diff = diff + (out[id] - tmIn)
        #print "Added", id
        #print "New Diff", diff
    else:
        dropped = dropped + 1
print "Dropped:", dropped
print "Total Diff of", diff, "over", num, "trials"

"""if len(out) != len(inp):
    nl = min(len(out)-5, len(inp)-5)
    nl = nl-(nl % 20)
    trunc(f1, nl)
    trunc(f2, nl)
    print "Truncated to len", nl
else:
    num = len(out)
    diff = datetime.timedelta(0)
    for ind in range(num):
        to = parse.parse(out[ind])
        ti = parse.parse(inp[ind])
        diff = diff + (to-ti)

    print "Total Diff of", diff, "over", num, "trials"
"""
