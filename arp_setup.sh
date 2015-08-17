#!/usr/bin/env bash

set -e

gcc -o ifaddrs getifaddrs.c
./ifaddrs | tee ipv4/arpv4/ips.static.orig > /dev/null
grep -e lo -e tap ipv4/arpv4/ips.static.orig | sed -e 's/10.0.0.2/10.0.0.1/g' > ipv4/ips.static
echo '00:34:45:de:ca:de' > ethernet/external_mac.static

(
	cd ipv4/arpv4
	cat ips.static.orig > ips_mac.static.orig
	python ips_mac_static_process.py < ips_mac.static.orig > ips_mac.static
)