#!/bin/bash

# run val1(genesis) node
docker-compose up -d val1

# wait for node to fully wakeup
sleep 2s

# get val1's tendermint node addr
val1addr=$(docker exec -it val1 tendermint show_node_id | tr -d '\015')

# update seed node's peer set with val1addr on docker-compose.yml 
sed -e s/__val1_addr__/$val1addr/ -i.tmp docker-compose.yml
