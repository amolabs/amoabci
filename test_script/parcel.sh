#!/bin/bash

#if [ $# != 0 ]; then
#	echo "usage: $(basename $0) <recipient_user>"
#	exit 0
#fi

. testaddr.sh

AMO1=1000000000000000000

STOID=1
STOURL="https://amo.foundation"

P1="000000017465737470617263656C6964"
CUSTODY="11ffeeff"

fail() {
	echo "test failed"
	echo $1
	exit -1
}

echo "faucet transfer coin to tu2: 1 AMO"
out=$($CLI tx --broadcast=commit transfer $CLIOPT --user tgenesis "$tu2" "$AMO1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "setup storage"
out=$($CLI tx --broadcast=commit setup $CLIOPT --user tu1 "$STOID" "$STOURL" 0 0 | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu1 register p1"
out=$($CLI tx --broadcast=commit register $CLIOPT --user tu1 "$P1" "$CUSTODY" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu1 discard p1"
out=$($CLI tx --broadcast=commit discard $CLIOPT --user tu1 "$P1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu1 register p1"
out=$($CLI tx --broadcast=commit register $CLIOPT --user tu1 "$P1" "$CUSTODY" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu2 request p1 with 1 AMO"
out=$($CLI tx --broadcast=commit request $CLIOPT --user tu2 "$tu2" "$P1" "$AMO1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu2 cancel p1 request"
out=$($CLI tx --broadcast=commit cancel $CLIOPT --user tu2 "$tu2" "$P1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu2 request p1 with 1 AMO"
out=$($CLI tx --broadcast=commit request $CLIOPT --user tu2 "$tu2" "$P1" "$AMO1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu1 grant tu2 on p1, collect 1 AMO"
out=$($CLI tx --broadcast=commit grant $CLIOPT --user tu1 "$tu2" "$P1" "$CUSTODY" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

echo "tu1 revoke grant given to tu2 on p1"
out=$($CLI tx --broadcast=commit revoke $CLIOPT --user tu1 "$tu2" "$P1" | sed 's/\^\@//g')
info=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['deliver_tx']['info']")
if [ -z "$info" -o "$info" != "ok" ]; then fail $out; fi

