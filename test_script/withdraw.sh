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
    echo "Retract $(bc <<< "$AMOUNT / $AMO1") AMO: del$i"
    docker exec -it testcli amocli tx retract $OPT --user tdel$i "$AMOUNT"
done

for ((i=FROM; i<=NODENUM; i++))
do
    echo "Withdraw $(bc <<< "$AMOUNT / $AMO1") AMO: val$i"

    # to prevent crash when no stake
    if [ "$i" -eq "$NODENUM" ]; then
        docker exec -it testcli amocli tx withdraw $OPT --user tval$i "$AMO1"
    else
        docker exec -it testcli amocli tx withdraw $OPT --user tval$i "$AMOUNT"
    fi
done

$ROOT/qb.sh "$NODENUM" "$OPT"
$ROOT/qs.sh "$NODENUM" "$OPT"
$ROOT/qd.sh "$NODENUM" "$OPT"

