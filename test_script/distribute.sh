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

$ROOT/qb.sh "$NODENUM"
. $ROOT/get_key.sh

for ((i=FROM; i<=NODENUM; i++))
do
    valaddr=tval$i

	echo "faucet to tval$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLIOPT tx transfer $OPT --user tgenesis "${!valaddr}" "$AMOUNT")
	h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
	if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi
    
    deladdr=tdel$i

	echo "faucet to tdel$i: $(bc <<< "$AMOUNT / $AMO1") AMO"
	out=$($CLIOPT tx transfer $OPT --user tgenesis "${!deladdr}" "$AMOUNT")
	h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
	if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi
done

echo "faucet to tu1: $(bc <<< "$AMOUNT / $AMO1") AMO"
out=$($CLIOPT tx transfer $OPT --user tgenesis "$tu1" "$AMOUNT")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "faucet to tu2: $(bc <<< "$AMOUNT / $AMO1") AMO"
out=$($CLIOPT tx transfer $OPT --user tgenesis "$tu2" "$AMOUNT")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

$ROOT/qb.sh "$NODENUM"

