#!/bin/bash

ROOT=$(dirname $0)

fail() {
	echo "test failed"
	echo $1
	exit -1
}

out=$($CLI tx stake $CLIOPT --user tu1 qQLA6Z0GDGYWzsxSbzQocmeFcCgSkc6cK9fY+M2YiOc= 10000)
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

