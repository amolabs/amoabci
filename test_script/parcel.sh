#!/bin/bash

#if [ $# == 0 ]; then
#	echo "usage: $(basename $0) <recipient_user>"
#	exit 0
#fi

. $(dirname $0)/get_key.sh

OPT=$1

AMO1=1000000000000000000

P1="7465737470617263656C6964"
CUSTODY="11ffeeff"

echo "faucet transfer coin to tu2: 1 AMO"
docker exec -it testcli amocli tx transfer $OPT --user tgenesis "$tu2" "$AMO1"

echo "tu1 register p1"
docker exec -it testcli amocli tx register $OPT --user tu1 "$P1" "$CUSTODY"

echo "query parcel info p1"
docker exec -it testcli amocli query parcel $OPT "$P1"

echo "tu1 discard p1"
docker exec -it testcli amocli tx discard $OPT --user tu1 "$P1"

echo "query parcel info p1"
docker exec -it testcli amocli query parcel $OPT "$P1"

echo "tu1 register p1"
docker exec -it testcli amocli tx register $OPT --user tu1 "$P1" "$CUSTODY" 

echo "query parcel info p1"
docker exec -it testcli amocli query parcel $OPT "$P1"

echo "tu2 request p1 with 1 AMO"
docker exec -it testcli amocli tx request $OPT --user tu2 "$P1" "$AMO1"

echo "query request of tu2 for p1"
docker exec -it testcli amocli query request $OPT "$tu2" "$P1" 

echo "query usage of tu2 for p1" 
docker exec -it testcli amocli query usage $OPT "$tu2" "$P1"

echo "tu1 grant tu2 on p1, collect 1 AMO"
docker exec -it testcli amocli tx grant $OPT --user tu1 "$P1" "$tu2" "$CUSTODY"

echo "query usage of tu2 for p1" 
docker exec -it testcli amocli query usage $OPT "$tu2" "$P1"

echo "tu1 revoke grant given to tu2 on p1"
docker exec -it testcli amocli tx revoke $OPT --user tu1 "$P1" "$tu2"

echo "query usage of tu2 for p1" 
docker exec -it testcli amocli query usage $OPT "$tu2" "$P1"

