#!/bin/bash

. $(dirname $0)/env.sh

echo "---- start"
amocli tx transfer --user t0 $t1 1000000000000000000
amocli tx transfer --user t0 $t2 1000000000000000000
amocli tx transfer --user t0 $u0 10000000000000
amocli tx transfer --user t0 $u1 10000000000000
amocli tx transfer --user t0 $u2 10000000000000
echo "---- end"

. $(dirname $0)/qb.sh

