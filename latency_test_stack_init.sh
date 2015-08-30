#!/usr/bin/env bash

pkill runStack
pkill tapip
#go build runStack.go
setcap CAP_NET_RAW=epi ./runStack
./runStack > /dev/null
