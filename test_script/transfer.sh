#!/bin/bash

if [ $# != 1 ]; then
	echo "usage: $(basename $0) <recipient_user>"
	exit 0
fi

. $(dirname $0)/env.sh

recpuser=$1
echo "recipient: $recpuser"
recp=${!recpuser}
echo "address: $recp"

. $(dirname $0)/qb.sh

echo "---- start"
if [ "$recpuser" != "t0" ]; then amocli tx transfer --user t0 $recp 1000; fi
if [ "$recpuser" != "t1" ]; then amocli tx transfer --user t1 $recp 1000; fi
if [ "$recpuser" != "t2" ]; then amocli tx transfer --user t2 $recp 1000; fi
if [ "$recpuser" != "u0" ]; then amocli tx transfer --user u0 $recp 1000; fi
if [ "$recpuser" != "u1" ]; then amocli tx transfer --user u1 $recp 1000; fi
if [ "$recpuser" != "u2" ]; then amocli tx transfer --user u2 $recp 1000; fi
echo "---- end"

. $(dirname $0)/qb.sh

