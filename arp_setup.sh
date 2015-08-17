#!/usr/bin/env bash

set -e

gcc -o ifaddrs getifaddrs.c
./ifaddrs > ipv4/arpv4/ips.static.orig
echo '00:34:45:de:ca:de' > ethernet/external_mac.static
echo '10.0.0.3' > ipv4/ipv4src/external_ip.static
echo '10.0.0.2' > ipv4/ipv4src/external_gateway.static

(
	cd ipv4/arpv4
	grep -e lo -e tap ips.static.orig > ips_mac.static.orig
	python ips_mac_static_process.py < ips_mac.static.orig > ips_mac.static
)
