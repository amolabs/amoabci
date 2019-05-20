#!/bin/bash

ROOT=$(dirname $0)

$ROOT/key.sh
$ROOT/distribute.sh

$ROOT/stake.sh

#$ROOT/transfer.sh t0
#$ROOT/transfer.sh t1
#$ROOT/transfer.sh t2

$ROOT/parcel.sh

$ROOT/withdraw.sh
