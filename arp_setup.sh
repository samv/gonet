#!/usr/bin/env bash

set -e

gcc -o ifaddrs getifaddrs.c
./ifaddrs > arp/static.config
