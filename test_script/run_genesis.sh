#!/bin/bash

docker-compose up --no-start val1
docker exec -it val1 mkdir -p /tendermint/config
docker exec -it val1 mkdir -p /tendermint/data

WD=$(dirname $0 | tr -d '\n')
docker cp $WD/priv_validator_key.json val1:/tendermint/config/
docker cp $WD/priv_validator_state.json val1:/tendermint/data/

# run val1(genesis) node
docker-compose up -d val1

# wait for node to fully wakeup
sleep 2s

# get val1's tendermint node addr
val1addr=$(docker exec -it val1 tendermint show_node_id | tr -d '\015' | tr -d '^@')

# update seed node's peer set with val1addr on docker-compose.yml 
sed -e s/__val1_addr__/$val1addr/ -i.tmp docker-compose.yml
