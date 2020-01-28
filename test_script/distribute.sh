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

$ROOT/qb.sh 
. testaddr.sh

for ((i=FROM; i<=NODENUM; i++)); do
    valaddr=tval$i

	echo "faucet to tval$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLI tx --broadcast=commit transfer $CLIOPT --user tgenesis "${!valaddr}" "$AMOUNT")
	info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
	if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi
    
    deladdr=tdel$i

	echo "faucet to tdel$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLI tx --broadcast=commit transfer $CLIOPT --user tgenesis "${!deladdr}" "$AMOUNT")
	info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
	if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi
done

echo "faucet to tu1: $(bc <<< "$AMOUNT / $AMO1") AMO"
out=$($CLI tx --broadcast=commit transfer $CLIOPT --user tgenesis "$tu1" "$AMOUNT")
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "faucet to tu2: $(bc <<< "$AMOUNT / $AMO1") AMO"
out=$($CLI tx --broadcast=commit transfer $CLIOPT --user tgenesis "$tu2" "$AMOUNT")
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

$ROOT/qb.sh 

