#!/bin/bash

NODENUM=$1

AMOUNT=1000

. $(dirname $0)/get_key.sh
. $(dirname $0)/qb.sh "$NODENUM"

for ((i=1; i<=NODENUM; i++))
do
    for ((j=1; j<=NODENUM; j++))
    do
        
        if [ "$i" -ne "$j" ]; then
            echo "tval$i -> tval$j: $AMOUNT"
            addr=tval$j
            $CLIOPT tx transfer $OPT --user tval$i ${!addr} "$AMOUNT" 
        fi

        echo "tval$i -> tdel$j: $AMOUNT"
        addr=tdel$j
        $CLIOPT tx transfer $OPT --user tval$i ${!addr} "$AMOUNT"

        if [ "$i" -ne "$j" ]; then
            echo "tdel$i -> tdel$j: $AMOUNT"
            $CLIOPT tx transfer $OPT --user tdel$i ${!addr} "$AMOUNT"
        fi

        echo "tdel$i -> tval$j: $AMOUNT"
        addr=tval$j
        $CLIOPT tx transfer $OPT --user tdel$i ${!addr} "$AMOUNT"

    done
    echo "tval$i, tdel$i -> tu1: $AMOUNT"
    addr=tu1
    $CLIOPT tx transfer $OPT --user tval$i ${!addr} "$AMOUNT"
    $CLIOPT tx transfer $OPT --user tdel$i ${!addr} "$AMOUNT"

    echo "tval$i, tdel$i -> tu2: $AMOUNT"
    addr=tu2
    $CLIOPT tx transfer $OPT --user tval$i ${!addr} "$AMOUNT"
    $CLIOPT tx transfer $OPT --user tdel$i ${!addr} "$AMOUNT"
done

. $(dirname $0)/qb.sh "$NODENUM"

