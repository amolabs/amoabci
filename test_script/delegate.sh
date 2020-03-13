#!/bin/bash

set -e

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
    addr=tval$i
    out=$(docker exec -it val$i amod tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")
	if [ $? -ne 0 ]; then fail $out; fi

	echo "delegate from tdel$i to tval$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLI tx --broadcast=commit delegate $CLIOPT --user tdel$i "${!addr}" "$AMOUNT")
	info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
	if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

done

$ROOT/qd.sh "$NODENUM"

