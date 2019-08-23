#!/bin/bash

# get amo keys
keys=$($CLIOPT key list)
sleep 2s

eval $(echo "$keys"| tr -d '\r' | awk '{ printf "%s=%s\n",$2,$4 }')
