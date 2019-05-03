#!/bin/bash

#if [ $# == 0 ]; then
#	echo "usage: $(basename $0) <recipient_user>"
#	exit 0
#fi

. $(dirname $0)/env.sh
. $(dirname $0)/qb.sh

echo "---- start"
amocli tx register --user u0 10ffe9 1f2faacc
amocli tx register --user u1 11ffe9 1f2faacc
amocli tx register --user u2 12ffe9 1f2faacc
echo "---- end"

. $(dirname $0)/qb.sh
amocli query parcel 10ffe9
amocli query parcel 11ffe9
amocli query parcel 12ffe9

echo "---- start"
amocli tx request --user t0 10ffe9 1000
amocli tx request --user t1 11ffe9 1000
amocli tx request --user t2 12ffe9 1000
echo "---- end"

. $(dirname $0)/qb.sh
amocli query request 10ffe9 $t0
amocli query request 11ffe9 $t1
amocli query request 12ffe9 $t2

echo "---- start"
amocli tx grant --user u0 10ffe9 $t0 1f1f1f1f
amocli tx grant --user u1 11ffe9 $t1 1f1f1f1f
amocli tx grant --user u2 12ffe9 $t2 1f1f1f1f
echo "---- end"

. $(dirname $0)/qb.sh
amocli query usage 10ffe9 $t0
amocli query usage 11ffe9 $t1
amocli query usage 12ffe9 $t2

