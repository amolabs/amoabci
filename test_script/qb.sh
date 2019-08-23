#!/bin/bash

NODENUM=$1

fail() {
	echo "test failed"
	echo $1
	exit -1
}

. $(dirname $0)/get_key.sh

out=$($CLIOPT query balance $OPT $tgenesis)
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of genesis: "$out

for ((i=1; i<=NODENUM; i++))
do
    addr=tval$i
	out=$($CLIOPT query balance $OPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
    echo "balance of tval$i: "$out 

    addr=tdel$i
	out=$($CLIOPT query balance $OPT ${!addr})
	if [ $? -ne 0 ]; then fail $out; fi
    echo "balance of tdel$i: "$out 
done

out=$($CLIOPT query balance $OPT "$tu1")
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of tu1: "$out 

out=$($CLIOPT query balance $OPT "$tu2")
if [ $? -ne 0 ]; then fail $out; fi
echo "balance of tu2: "$out 

