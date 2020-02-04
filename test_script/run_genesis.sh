#!/bin/bash

get_block_height() {
	out=$($CLI query node)
	height=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['sync_info']['latest_block_height']")

	echo "$height"
}

check_block_height() {
	expected_height=$1
	height=$(get_block_height)

	while (( expected_height > height )); do
		height=$(get_block_height)
		sleep 1
	done
}

check_rpc_status() {
	printf "wait for val1 node(entry point) to fully wake up "
	until $($CURL --output /dev/null --silent --head --fail $CURLOPT); do
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

echo "wait for block progresses sufficiently"
check_block_height "2"

echo "update config set of amocli with the one from amo node"
out=$($CLI query node $CLIOPT)
out=$($CLI query config $CLIOPT)
