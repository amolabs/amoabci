#!/bin/bash

. $(dirname $0)/env.sh

val0key=$(docker exec -it seed tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")
val1key=$(docker exec -it val1 tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")
val2key=$(docker exec -it val2 tendermint show_validator | python -c "import sys, json; print json.load(sys.stdin)['value']")

. $(dirname $0)/qb.sh
. $(dirname $0)/qs.sh

echo "seed val key:" $val0key
echo "val1 val key:" $val1key
echo "val2 val key:" $val2key

echo "---- start"
amocli tx stake --json --user t0 "$val0key" 1000000000000000000
amocli tx stake --json --user t1 "$val1key" 1000000000000000000
amocli tx stake --json --user t2 "$val2key" 1000000000000000000
amocli tx delegate --json --user d0 $t0 10000000000000
amocli tx delegate --json --user d1 $t1 10000000000000
amocli tx delegate --json --user d2 $t2 10000000000000
echo "---- end"

. $(dirname $0)/qb.sh
. $(dirname $0)/qs.sh

