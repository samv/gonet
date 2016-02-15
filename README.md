A network stack written in Go.

To generate the logs run `git log --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%ci) %C(bold blue)<%an>%Creset' --abbrev-commit | tac | python -c 'import sys; print "\n".join([x.strip("\n")[:15].replace("/", "lolxxx").replace("\\", "/").replace("lolxxx", "\\") + x.strip("\n")[15:] for x in sys.stdin.readlines()])' | aha > log.htm`