#!/bin/bash

ROOT=$(dirname $0)

FROM=$1
NODENUM=$2
AMOUNT=$3
OPT=$4

AMO1=1000000000000000000

$ROOT/qb.sh "$NODENUM" "$OPT"

. $ROOT/get_key.sh

for ((i=FROM; i<=NODENUM; i++))
do
    valaddr=tval$i

    echo "Transfer $(bc <<< "$AMOUNT / $AMO1") AMO: tval$i"
    docker exec -it cli amocli tx transfer $OPT --user tgenesis "${!valaddr}" "$AMOUNT"
    
    deladdr=tdel$i

    echo "Transfer $(bc <<< "$AMOUNT / $AMO1") AMO: tdel$i"
    docker exec -it cli amocli tx transfer $OPT --user tgenesis "${!deladdr}" "$AMOUNT"
done

echo "Transfer $(bc <<< "$AMOUNT / $AMO1") AMO: tu1"
docker exec -it cli amocli tx transfer $OPT --user tgenesis "$tu1" "$AMOUNT"

echo "Transfer $(bc <<< "$AMOUNT / $AMO1") AMO: tu2"
docker exec -it cli amocli tx transfer $OPT --user tgenesis "$tu2" "$AMOUNT"

$ROOT/qb.sh "$NODENUM" "$OPT"

