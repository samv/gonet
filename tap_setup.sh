#!/usr/bin/env bash

#sudo ip link set dev tap0 up
#sudo ip addr add 10.0.0.1/24 dev tap0

if [ -d "~/tapip-primes" ]; then
  echo 'exit' | sudo ~/tapip-primes/tapip
else
  set -e
  git clone https://hsheth2@bitbucket.org/primesteam/tapip.git
  cd tapip
  make
  echo 'exit' | sudo ./tapip
  cd ..
  rm -rf tapip/
fi
