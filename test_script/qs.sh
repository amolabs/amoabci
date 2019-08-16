#!/bin/bash

NODENUM=$1
OPT=$2

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
    echo "stake of tval$i:" $(amocli query stake "$OPT" ${!addr})
done

