#!/usr/bin/env bash

pkill runStack
pkill tapip
cd ../../../tapip
sudo ./tapip > /dev/null
