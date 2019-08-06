#!/bin/bash

# This script is for bootstrapping docker containers(node) sequentially

# node composition (6)  
#
# 0. seed (non-validator)
# 1. val1 : genesis validator
# 2. val2 : validator
# 3. val3 : validator
# 4. val4 : validator
# 5. val5 : validator
# 6. val6 : validator

# node dependency
#
# val1 <- seed
# seed <- val2, val3, val4, val5, val6

if [ ! -f docker-compose.yml.in ]; then
    echo "docker-compose.yml.in doesn't exist"
    exit
fi

DATAROOT=$HOME/.amotest
VALNUM=6

AMO100=100000000000000000000

# build docker image
docker build -t amolabs/amotest DOCKER

# set basic environments
cp -f docker-compose.yml.in docker-compose.yml
sed -e s#@dataroot@#$DATAROOT# -i docker-compose.yml

mkdir -p $DATAROOT/seed/amo
mkdir -p $DATAROOT/seed/tendermint/config
mkdir -p $DATAROOT/seed/tendermint/data

for ((i=1; i<=VALNUM; i++))
do
    mkdir -p $DATAROOT/val$i/amo
    mkdir -p $DATAROOT/val$i/tendermint/config
    mkdir -p $DATAROOT/val$i/tendermint/data
   
    # generate validator's amo key
    amocli key remove val$i
    amocli key generate val$i --encrypt=false
done

# get validators' amo address
eval $(amocli key list | awk '{ if ($2 != "t0") printf "%s=%s\n",$2,$4 }')

# run val1(genesis) node
docker-compose up -d val1

# get val1's tendermint validator key
val1pubkey=$(docker exec -it val1 tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")

# get val1's tendermint node addr
val1addr=$(docker exec -it val1 tendermint show_node_id | tr -d '\015')

# update seed node's peer set with val1addr on docker-compose.yml 
sed -e s/@val1_addr@/$val1addr/ -i docker-compose.yml

# faucet to val1 owner: 100 AMO
echo "Transfer 100 AMO: genesis -> val1"
amocli tx transfer --json --user genesis "$val1" "$AMO100"

# stake for val1
echo "Stake 100 AMO: val1"
amocli tx stake --json --user val1 "$val1pubkey" "$AMO100"

# wait for val1 to fully wakeup
sleep 2s

# run seed node
docker-compose up -d seed

# get val1's tendermint node addr
seedaddr=$(docker exec -it seed tendermint show_node_id | tr -d '\015')

# update seed node's peer set with val1addr on docker-compose.yml 
sed -e s/@seed_addr@/$seedaddr/ -i docker-compose.yml

# wait for seed to fully wakeup
sleep 2s

# val nodes: val2, val3, val4, val5, val6
for ((i=2; i<=VALNUM; i++))
do
    docker-compose up -d val$i
done

# faucet to the validator owners: 100 AMO for each
for ((i=2; i<=VALNUM; i++))
do
    tmpaddr=val$i

    echo "Transfer 100 AMO: genesis -> val$i"
    amocli tx transfer --json --user genesis "${!tmpaddr}" "$AMO100"
done

# stake for val2, val3, val4, val5, val6
for ((i=2; i<=VALNUM; i++))
do
    tmppubkey=$(docker exec -it val$i tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")
    
    echo "Stake 100 AMO: val$i"
    amocli tx stake --json --user val$i "$tmppubkey" "$AMO100"
done

