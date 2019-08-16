#!/bin/bash

ROOT=$(dirname $0)

FROM=$1
NODENUM=$2
AMOUNT=$3
OPT=$4

AMO1=1000000000000000000

$ROOT/qb.sh "$NODENUM" "$OPT"
$ROOT/qs.sh "$NODENUM" "$OPT"
$ROOT/qd.sh "$NODENUM" "$OPT"

. $ROOT/get_key.sh

for ((i=FROM; i<=NODENUM; i++))
do

    addr=tval$i
    pubkey=$(docker exec -it val$i tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")

    echo "Stake $(bc <<< "$AMOUNT / $AMO1") AMO: val$i"
    amocli tx stake "$OPT" --json --user tval$i "$pubkey" "$AMOUNT"

    echo "Delegate $(bc <<< "$AMOUNT / $AMO1") AMO: del$i -> val$i"
    amocli tx delegate "$OPT" --json --user tdel$i "${!addr}" "$AMOUNT"
done

$ROOT/qb.sh "$NODENUM" "$OPT"
$ROOT/qs.sh "$NODENUM" "$OPT"
$ROOT/qd.sh "$NODENUM" "$OPT"

