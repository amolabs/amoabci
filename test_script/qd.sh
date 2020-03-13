#!/bin/bash

set -e

NODENUM=$1

fail() {
	echo "test failed"
	echo $1
	exit -1
}

. testaddr.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tdel$i
	out=$($CLI query delegate $CLIOPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
	echo "delegate of tdel$i: "$out
done

