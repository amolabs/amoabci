#!/bin/bash

ROOT=$(dirname $0)

. $ROOT/key.sh
. $ROOT/distribute.sh

. $ROOT/stake.sh

. $ROOT/transfer.sh t0
. $ROOT/transfer.sh t1
. $ROOT/transfer.sh t2

. $ROOT/unstake.sh

. $ROOT/stake.sh

. $ROOT/transfer.sh u0
. $ROOT/transfer.sh u1
. $ROOT/transfer.sh u2

. $ROOT/withdraw.sh
