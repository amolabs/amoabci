#!/bin/bash

NODENUM=$1

fail() {
	echo "test failed"
	echo $1
	exit -1
}

. $(dirname $0)/get_key.sh

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
	out=$($CLIOPT query stake $OPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
	echo "stake of tval$i: "$out
done

