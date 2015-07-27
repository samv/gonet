#!/usr/bin/env bash

echo "Formatting Files:"
goimports -l -w ./
echo "Finished Formatting"
