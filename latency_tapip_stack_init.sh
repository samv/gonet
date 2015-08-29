#!/usr/bin/env bash

pkill runStack
pkill tapip
cd ../../../tapip
./tapip > /dev/null 2>&1
