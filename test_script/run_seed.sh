#!/bin/bash

# run seed node
docker-compose up -d seed

# wait for node to fully wakeup
sleep 2s

# get val1's tendermint node addr
seedaddr=$(docker exec -it seed tendermint show_node_id | tr -d '\015')

# update validator nodes' peer set with val1addr on docker-compose.yml 
sed -e s/__seed_addr__/$seedaddr/ -i.tmp docker-compose.yml
