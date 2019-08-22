#!/bin/bash

#if [ $# == 0 ]; then
#	echo "usage: $(basename $0) <recipient_user>"
#	exit 0
#fi

. $(dirname $0)/get_key.sh

AMO1=1000000000000000000

P1="7465737470617263656C6964"
CUSTODY="11ffeeff"

fail() {
	echo "test failed"
	echo $1
	exit -1
}

echo "faucet transfer coin to tu2: 1 AMO"
out=$($CLIOPT tx transfer $OPT --user tgenesis "$tu2" "$AMO1")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu1 register p1"
out=$($CLIOPT tx register $OPT --user tu1 "$P1" "$CUSTODY")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu1 discard p1"
out=$($CLIOPT tx discard $OPT --user tu1 "$P1")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu1 register p1"
out=$($CLIOPT tx register $OPT --user tu1 "$P1" "$CUSTODY")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu2 request p1 with 1 AMO"
out=$($CLIOPT tx request $OPT --user tu2 "$P1" "$AMO1")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu1 grant tu2 on p1, collect 1 AMO"
out=$($CLIOPT tx grant $OPT --user tu1 "$P1" "$tu2" "$CUSTODY")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

echo "tu1 revoke grant given to tu2 on p1"
out=$($CLIOPT tx revoke $OPT --user tu1 "$P1" "$tu2")
h=$(echo $out | python -c "import sys, json; print json.load(sys.stdin)['height']")
if [ -z "$h" -o "$h" == "0" ]; then fail $out; fi

