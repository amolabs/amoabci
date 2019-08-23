#!/bin/bash

fail() {
	echo "test failed"
	echo $1
	exit -1
}

echo "run seed node"
out=$(docker-compose up --no-start seed)
if [ $? -ne 0 ]; then fail $out; fi

out=$(docker-compose run --rm seed mkdir -p /tendermint/config)
if [ $? -ne 0 ]; then fail $out; fi

WD=$(dirname $0)

out=$(docker cp $WD/genesis.json seed:/tendermint/config/)
if [ $? -ne 0 ]; then fail $out; fi

out=$(docker-compose up -d seed)
if [ $? -ne 0 ]; then fail $out; fi

echo "wait for node to fully wakeup"
sleep 1s

echo "get val1's tendermint node addr"
out=$(docker exec -it seed tendermint show_node_id | tr -d '\015' | tr -d '^@')
if [ $? -ne 0 ]; then fail $out; fi

echo "update validator nodes' peer set with val1addr on docker-compose.yml"
out=$(sed -e s/__seed_addr__/$out/ -i.tmp docker-compose.yml)
if [ $? -ne 0 ]; then fail $out; fi

