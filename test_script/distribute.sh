#!/bin/bash

. $(dirname $0)/env.sh

echo "---- start"
amocli tx transfer --json --user t0 $t1 1000000000000000000
amocli tx transfer --json --user t0 $t2 1000000000000000000
amocli tx transfer --json --user t0 $d0 10000000000000
amocli tx transfer --json --user t0 $d1 10000000000000
amocli tx transfer --json --user t0 $d2 10000000000000
echo "---- end"

. $(dirname $0)/qb.sh

