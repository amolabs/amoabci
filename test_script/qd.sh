#!/bin/bash

NODENUM=$1

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tdel$i
    echo "stake of tdel$i:" $(amocli query delegate ${!addr})
done

