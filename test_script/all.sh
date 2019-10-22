#!/bin/bash

################################################################################
# node composition (6) # 
########################
# 0. seed (non-validator)
# 1. val1 : genesis validator
# 2. val2 : validator
# 3. val3 : validator
# 4. val4 : validator
# 5. val5 : validator
# 6. val6 : validator
#
################################################################################
# node dependency #
###################
# val1 <- seed
# seed <- val2, val3, val4, val5, val6
################################################################################

if [ ! -f docker-compose.yml.in ]; then
    echo "docker-compose.yml.in doesn't exist"
    exit
fi
cp -f docker-compose.yml.in docker-compose.yml

export CLI=amocli
export CLIOPT="--json"
export CURL=curl
export CURLOPT="localhost:26657" 

ROOT=$(dirname $0)
DATAROOT=$PWD/testdata
NODENUM=6

AMO100=100000000000000000000

echo "build docker image"
make docker

echo "generate key set"
$ROOT/gen_key.sh "$NODENUM"

echo "set basic environments"
$ROOT/env.sh "$DATAROOT" "$NODENUM"

echo "bootstrap nodes" 
$ROOT/bootstrap.sh "$NODENUM"

echo "distribute coins"
$ROOT/distribute.sh 1 "$NODENUM" "$AMO100"

echo "parcel related transactions"
$ROOT/parcel.sh

$ROOT/withdraw.sh 1 "$NODENUM" "$AMO100"
