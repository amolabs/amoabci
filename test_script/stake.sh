#!/bin/bash

ROOT=$(dirname $0)

FROM=$1
NODENUM=$2
AMOUNT=$3

AMO1=1000000000000000000

$ROOT/qb.sh "$NODENUM" 
$ROOT/qs.sh "$NODENUM"
$ROOT/qd.sh "$NODENUM"

. $ROOT/get_key.sh

for ((i=FROM; i<=NODENUM; i++))
do

    addr=tval$i
    pubkey=$(docker exec -it val$i tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")

    echo "Stake $(bc <<< "$AMOUNT / $AMO1") AMO: val$i"
    amocli tx stake --json --user tval$i "$pubkey" "$AMOUNT"

    echo "Delegate $(bc <<< "$AMOUNT / $AMO1") AMO: del$i -> val$i"
    amocli tx delegate --json --user tdel$i "${!addr}" "$AMOUNT"
done

$ROOT/qb.sh "$NODENUM"
$ROOT/qs.sh "$NODENUM"
$ROOT/qd.sh "$NODENUM"

