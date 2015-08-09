#!/usr/bin/env bash

set -e

gcc -o ifaddrs getifaddrs.c
./ifaddrs > arpv4/static.config
