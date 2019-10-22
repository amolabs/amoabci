#!/bin/bash

NODENUM=$1

check_docker_status() {
	printf "wait for %s	to fully wake up " $1
	until [ $(docker inspect -f {{.State.Running}} $1) == "true" ]; do
		printf "."
		sleep 0.1
	done
	printf " it is fully up!\n"
}

fail() {
	echo "test failed"
	echo $1
	exit -1
}

# val nodes: val2, val3, val4, val5, val6
for ((i=2; i<=NODENUM; i++))
do
    echo "run val$i node"
	out=$(docker-compose up --no-start val$i)
	if [ $? -ne 0 ]; then fail $out; fi

	out=$(docker-compose run --rm val$i mkdir -p /tendermint/config)
	if [ $? -ne 0 ]; then fail $out; fi

	WD=$(dirname $0)
	
	out=$(docker cp $WD/genesis.json val$i:/tendermint/config/)
	if [ $? -ne 0 ]; then fail $out; fi

	out=$(docker-compose up -d val$i)
	if [ $? -ne 0 ]; then fail $out; fi

	check_docker_status "val$i"
done
