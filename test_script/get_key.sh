#!/bin/bash

# get amo keys
eval $(docker exec -it cli amocli key list | awk '{ printf "%s=%s\n",$2,$4 }')
