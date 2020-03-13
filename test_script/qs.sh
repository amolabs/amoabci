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
    addr=tval$i
	out=$($CLI query stake $CLIOPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
	echo "stake of tval$i: "$out
done

