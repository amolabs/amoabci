#!/bin/bash

# get amo keys
eval $(amocli key list | awk '{ printf "%s=%s\n",$2,$4 }')
