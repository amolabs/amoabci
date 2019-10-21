#!/bin/bash

check_rpc_status() {
	printf "wait for val1 node(entry point) to fully wake up "
	until $(curl --output /dev/null --silent --head --fail http://localhost:26657); do
    	printf "."
    	sleep 0.5 
	done
	printf " it is fully up!\n"
}

fail() {
	echo "test failed"
	echo $1
	exit -1
}

echo "run val1(genesis) node"
out=$(docker-compose up --no-start val1)
if [ $? -ne 0 ]; then fail $out; fi

out=$(docker-compose run --rm val1 mkdir -p /tendermint/config)
if [ $? -ne 0 ]; then fail $out; fi
out=$(docker-compose run --rm val1 mkdir -p /tendermint/data)
if [ $? -ne 0 ]; then fail $out; fi

WD=$(dirname $0)

out=$(docker cp $WD/genesis.json val1:/tendermint/config/)
if [ $? -ne 0 ]; then fail $out; fi
out=$(docker cp $WD/priv_validator_key.json val1:/tendermint/config/)
if [ $? -ne 0 ]; then fail $out; fi
out=$(docker cp $WD/priv_validator_state.json val1:/tendermint/data/)
if [ $? -ne 0 ]; then fail $out; fi

out=$(docker-compose up -d val1)
if [ $? -ne 0 ]; then fail $out; fi

check_rpc_status

echo "get val1's tendermint node addr"
out=$(docker exec -it val1 tendermint show_node_id | tr -d '\015' | tr -d '^@')
if [ $? -ne 0 ]; then fail $out; fi

echo "update seed node's peer set with val1addr on docker-compose.yml"
out=$(sed -e s/__val1_addr__/$out/ -i.tmp docker-compose.yml)
if [ $? -ne 0 ]; then fail $out; fi

