#!/bin/bash

NODENUM=$1

. $(dirname $0)/get_key.sh

echo "balance of genesis:"$(amocli query balance $tgenesis)

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
    echo "balance of tval$i:" $(amocli query balance ${!addr})
done

