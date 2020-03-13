#!/bin/bash

set -e

NODENUM=$1

fail() {
	echo "test failed"
	echo $1
	exit -1
}

. testaddr.sh

out=$($CLI query balance $CLIOPT $tgenesis)
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of genesis: "$out

for ((i=1; i<=NODENUM; i++)); do
    addr=tval$i
	out=$($CLI query balance $CLIOPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
    echo "balance of tval$i: "$out 

    addr=tdel$i
	out=$($CLI query balance $CLIOPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
    echo "balance of tdel$i: "$out 
done

out=$($CLI query balance $CLIOPT "$tu1")
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of tu1: "$out 

out=$($CLI query balance $CLIOPT "$tu2")
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of tu2: "$out 

