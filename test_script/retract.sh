#!/bin/bash

ROOT=$(dirname $0)

FROM=$1
NODENUM=$2
AMOUNT=$3

AMO1=1000000000000000000

fail() {
	echo "test failed"
	echo $1
	exit -1
}

$ROOT/qd.sh "$NODENUM"

. testaddr.sh

for ((i=FROM; i<=NODENUM; i++))
do
    echo "retract del$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLI tx --broadcast=commit retract $CLIOPT --user tdel$i "$AMOUNT")
	info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
	if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi
done

$ROOT/qd.sh "$NODENUM"
