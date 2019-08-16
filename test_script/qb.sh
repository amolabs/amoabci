#!/bin/bash

NODENUM=$1
OPT=$2

. $(dirname $0)/get_key.sh

echo "balance of genesis:"$(amocli query balance "$OPT" $tgenesis)

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
    echo "balance of tval$i:" $(amocli query balance "$OPT" ${!addr})
done

