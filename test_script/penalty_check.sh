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
out=$($CLI query stake $CLIOPT ${!addr})
echo "stake of tu1: "$out

a=$(echo $out | tr -d '^@' | python -c "import sys, json; print json.load(sys.stdin)['amount']")
if [ -z "$a" -o "$a" != "1" ]; then fail $out; fi
