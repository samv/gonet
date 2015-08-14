#!/usr/bin/env bash

set -e

gcc -o ifaddrs getifaddrs.c
./ifaddrs | tee ethernet/mac.static.orig ipv4/arpv4/ips.static > /dev/null
grep -e lo -e tap ipv4/arpv4/ips.static > ipv4/ips.static
(
	cd ethernet
	python static_process.py < mac.static.orig > mac.static
)