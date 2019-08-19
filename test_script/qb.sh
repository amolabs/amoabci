#!/bin/bash

NODENUM=$1

. $(dirname $0)/get_key.sh

echo "balance of genesis:"$($CLIOPT query balance $OPT $tgenesis)

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
    echo "balance of tval$i:" $($CLIOPT query balance $OPT ${!addr})
done

