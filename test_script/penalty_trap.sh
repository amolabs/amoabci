#!/bin/bash

set -e

ROOT=$(dirname $0)

fail() {
	echo "test failed"
	echo $1
	exit -1
}

. testaddr.sh
addr=tu1
val="qQLA6Z0GDGYWzsxSbzQocmeFcCgSkc6cK9fY+M2YiOc="

out=$($CLI query stake $CLIOPT ${!addr})
if [ $? -ne 0 ]; then fail $out; fi
echo "stake of tu1: "$out

echo "stake to tu1: 10000 MOTE"
out=$($CLI tx --broadcast=commit stake $CLIOPT --user tu1 $val 10000)
info=$(echo $out | tr -d '^@' | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

out=$($CLI query stake $CLIOPT ${!addr})
if [ $? -ne 0 ]; then fail $out; fi
echo "stake of tu1: "$out
