#!/bin/bash

. $(dirname $0)/env.sh

. $(dirname $0)/qb.sh
. $(dirname $0)/qs.sh

echo "---- start"
amocli tx retract --json --user d0 10000000000000
amocli tx retract --json --user d1 10000000000000
amocli tx retract --json --user d2 10000000000000
amocli tx withdraw --json --user t1 1000000000000000000
amocli tx withdraw --json --user t2 1000000000000000000
amocli tx withdraw --json --user t0 999999999999999999
echo "---- end"

. $(dirname $0)/qb.sh
. $(dirname $0)/qs.sh

