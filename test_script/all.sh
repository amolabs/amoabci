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

LOCKUP=$(cat $PWD/test_script/genesis.json | python -c "import sys, json; print json.load(sys.stdin)['app_state']['config']['lockup_period']")

get_block_height() {
	out=$($CLI status)
	height=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['sync_info']['latest_block_height']")

	echo "$height"
}

check_block_height() {
	stake_height=$1
	height=$(get_block_height)

	echo "stake=$stake_height, lockup=$LOCKUP, height=$height"

	while (( stake_height + LOCKUP - 1 > height )); do
		height=$(get_block_height)
		printf "."
		sleep 1
	done
	printf " reached at $height\n"
}

echo "build docker image"
make docker

echo "generate key set"
$ROOT/gen_key.sh "$NODENUM"

echo "set basic environments"
$ROOT/env.sh "$DATAROOT" "$NODENUM"

echo "bootstrap nodes" 
$ROOT/bootstrap.sh "$NODENUM"

stake_height=$(get_block_height)

echo "[should FAIL] withdraw staked coins for val2, val3, val4, val5, val6"
$ROOT/withdraw.sh 2 "$NODENUM" "$AMO100" "stake locked"

echo "delegate coins"
$ROOT/delegate.sh 1 "$NODENUM" "$AMO100"

echo "parcel related transactions"
$ROOT/parcel.sh

echo "retract delegated coins"
$ROOT/retract.sh 1 "$NODENUM" "$AMO100"

echo "wait for stakes locked at $stake_height to get unlocked"
check_block_height "$stake_height"

echo "[shoud SUCCEED] withdraw staked coins for val2, val3, val4, val5, val6"
$ROOT/withdraw.sh 2 "$NODENUM" "$AMO100" "ok"

echo "tu1's stake SHOULD be 1 mote"
$ROOT/penalty_check.sh
