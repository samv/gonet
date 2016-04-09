#!/usr/bin/env bash

#sudo ip link set dev tap0 up
#sudo ip addr add 10.0.0.1/24 dev tap0

if [ -d "~/tapip-primes" ]; then
  echo 'exit' | sudo ~/tapip-primes/tapip
else
  if [ ! -d "tapip" ]; then
      git clone https://github.com/hsheth2/tapip.git
  fi
  cd tapip
  make
  set -e
  echo 'exit' | sudo ./tapip
fi
