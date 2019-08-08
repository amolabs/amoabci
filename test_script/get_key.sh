#!/bin/sh

# get amo keys
eval $(amocli key list | awk '{ if ($2 != "t0") printf "%s=%s\n",$2,$4 }')
