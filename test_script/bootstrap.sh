#!/bin/bash

set -e

# This script is for bootstrapping docker containers(node) sequentially

NODENUM=$1
ROOT=$(dirname $0)

AMO100=100000000000000000000

echo "bootstrap genesis node"
$ROOT/run_genesis.sh

echo "faucet to val1 owner: 100 AMO"
$ROOT/distribute.sh 1 1 "$AMO100"

echo "stake for val1"
$ROOT/stake.sh 1 1 "$AMO100"

echo "tu1 stake to non-existing validator for downtime(lazy validstors) penalty test"
$ROOT/penalty_trap.sh

echo "bootstrap seed node"
$ROOT/run_seed.sh

echo "bootstrap validator nodes: val2, val3, val4, val5, val6"
$ROOT/run_validators.sh "$NODENUM"

echo "faucet to the validator owners: 100 AMO for each"
$ROOT/distribute.sh 2 "$NODENUM" "$AMO100"

echo "stake for val2, val3, val4, val5, val6"
$ROOT/stake.sh 2 "$NODENUM" "$AMO100"

echo "remove tmp files"
rm -f *.tmp
