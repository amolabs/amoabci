#!/bin/bash

NODENUM=$1
OPT=$2

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tdel$i
    echo "delegate of tdel$i:" $(docker exec -it testcli amocli query delegate $OPT ${!addr})
done

