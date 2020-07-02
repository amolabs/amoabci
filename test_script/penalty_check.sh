#!/bin/bash

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

# NOTE: this will work on circleci only
a=$(echo $out | tr -d '^@' | python -c "import sys, json; print json.load(sys.stdin)['amount']")
if [ -z "$a" -o "$a" != "1280000" ]; then fail $out; fi
