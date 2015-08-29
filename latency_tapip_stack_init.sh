#!/usr/bin/env bash

pkill runStack
pkill tapip
../../../tapip/tapip > /dev/null 2>&1 &
sleep 0.5
pkill tapip
echo "stall" | ../../../tapip/tapip > /dev/null 2>&1
