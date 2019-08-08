#!/bin/bash

#if [ $# == 0 ]; then
#	echo "usage: $(basename $0) <recipient_user>"
#	exit 0
#fi

. $(dirname $0)/get_key.sh

AMO1=1000000000000000000

P1="7465737470617263656C6964"
CUSTODY="11ffeeff"

echo "faucet transfer coin to tu2: 1 AMO"
amocli tx transfer --json --user tgenesis "$tu2" "$AMO1"

echo "tu1 register p1"
amocli tx register --json --user tu1 "$P1" "$CUSTODY"

echo "query parcel info p1"
amocli query parcel --json "$P1"

echo "tu1 discard p1"
amocli tx discard --json --user tu1 "$P1"

echo "query parcel info p1"
amocli query parcel --json "$P1"

echo "tu1 register p1"
amocli tx register --json --user tu1 "$P1" "$CUSTODY" 

echo "query parcel info p1"
amocli query parcel --json "$P1"

echo "tu2 request p1 with 1 AMO"
amocli tx request --json --user tu2 "$P1" "$AMO1"

echo "query request of tu2 for p1"
amocli query request --json "$tu2" "$P1" 

echo "query usage of tu2 for p1" 
amocli query usage --json "$tu2" "$P1"

echo "tu1 grant tu2 on p1, collect 1 AMO"
amocli tx grant --json --user tu1 "$P1" "$tu2" "$CUSTODY"

echo "query usage of tu2 for p1" 
amocli query usage --json "$tu2" "$P1"

echo "tu1 revoke grant given to tu2 on p1"
amocli tx revoke --json --user tu1 "$P1" "$tu2"

echo "query usage of tu2 for p1" 
amocli query usage --json "$tu2" "$P1"

