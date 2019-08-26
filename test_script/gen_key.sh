#!/bin/bash

# < key set >
#
# 01. tgenesis
# 02. tval1
# 03. tval2
# 04. tval3
# 05. tval4
# 06. tval5
# 07. tval6

# 08. tdel1
# 09. tdel2
# 10. tdel3
# 11. tdel4
# 12. tdel5
# 13. tdel6

# 14. tu1
# 15. tu2

GENESISPRIVKEY="McFS24Dds4eezIfe+lfoni02J7lfs2eQQyhwF51ufmA="

NODENUM=$1

fail() {
	echo "test failed"
	echo $1
	exit -1
}

echo "regenerate genesis key"
out=$($CLIOPT key remove tgenesis)
out=$($CLIOPT key import --username=tgenesis --encrypt=false "$GENESISPRIVKEY")
if [ $? -ne 0 ]; then fail "$out"; fi

for ((i=1; i<=NODENUM; i++))
do
	echo "regenerate tval$i key"
	out=$($CLIOPT key remove tval$i)
	out=$($CLIOPT key generate tval$i --encrypt=false)
	if [ $? -ne 0 ]; then fail "$out"; fi
	
	echo "regenerate tdel$i key"
	out=$($CLIOPT key remove tdel$i)
	out=$($CLIOPT key generate tdel$i --encrypt=false)
	if [ $? -ne 0 ]; then fail "$out"; fi
done

echo "regenerate tu1 key"
out=$($CLIOPT key remove tu1)
out=$($CLIOPT key generate tu1 --encrypt=false)
if [ $? -ne 0 ]; then fail "$out"; fi

echo "regenerate tu2 key"
out=$($CLIOPT key remove tu2)
out=$($CLIOPT key generate tu2 --encrypt=false)
if [ $? -ne 0 ]; then fail "$out"; fi

keys=$($CLIOPT key list)
echo "$keys"| tr -d '\r' | awk '{ printf "%s=%s\n",$2,$4 }' > testaddr.sh

