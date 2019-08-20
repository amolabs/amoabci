#!/bin/bash

NODENUM=$1

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
    echo "stake of tval$i:" $($CLIOPT query stake $OPT ${!addr})
done

