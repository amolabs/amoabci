#!/bin/bash

# get amo keys
eval $($CLIOPT key list | awk '{ printf "%s=%s\n",$2,$4 }')