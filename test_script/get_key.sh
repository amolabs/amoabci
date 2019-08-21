#!/bin/bash

# get amo keys
eval $($CLIOPT key list | tr -d '\r' | awk '{ printf "%s=%s\n",$2,$4 }')

