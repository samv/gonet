#!/usr/bin/env bash

set -e

echo '00:34:45:de:ca:de' > ethernet/external_mac.static
echo '10.0.0.3' > ipv4/external_ip.static
echo '10.0.0.2' > ipv4/external_gateway.static

sudo sysctl net.ipv4.ip_forward=1
sudo iptables -t nat -A PREROUTING -p tcp -d `/sbin/ifconfig eth0 | grep 'inet addr:' | cut -d: -f2 | awk '{ print $1}'` --dport 80 -j DNAT --to 10.0.0.3:80
sudo iptables -t nat -A POSTROUTING -j MASQUERADE
