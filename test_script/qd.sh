#!/bin/bash

NODENUM=$1

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tdel$i
    echo "delegate of tdel$i:" $($CLIOPT query delegate $OPT ${!addr})
done

