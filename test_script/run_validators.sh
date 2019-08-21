#!/bin/bash

NODENUM=$1

# val nodes: val2, val3, val4, val5, val6
for ((i=2; i<=NODENUM; i++))
do
    docker-compose up --no-start val$i
	docker-compose run --rm val$i mkdir -p /tendermint/config
	WD=$(dirname $0)
	docker cp $WD/genesis.json val$i:/tendermint/config/
    docker-compose up -d val$i
    
    # wait for node to fully wakeup
    sleep 1s
done
