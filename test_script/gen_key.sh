#!/bin/bash

set -e

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
out=$($CLI key remove tgenesis)
out=$($CLI key import --username=tgenesis --encrypt=false "$GENESISPRIVKEY")
if [ $? -ne 0 ]; then fail "$out"; fi

for ((i=1; i<=NODENUM; i++))
do
	echo "regenerate tval$i key"
	out=$($CLI key remove tval$i)
	out=$($CLI key generate tval$i --encrypt=false)
	if [ $? -ne 0 ]; then fail "$out"; fi
	
	echo "regenerate tdel$i key"
	out=$($CLI key remove tdel$i)
	out=$($CLI key generate tdel$i --encrypt=false)
	if [ $? -ne 0 ]; then fail "$out"; fi
done

echo "regenerate tu1 key"
out=$($CLI key remove tu1)
out=$($CLI key generate tu1 --encrypt=false)
if [ $? -ne 0 ]; then fail "$out"; fi

echo "regenerate tu2 key"
out=$($CLI key remove tu2)
out=$($CLI key generate tu2 --encrypt=false)
if [ $? -ne 0 ]; then fail "$out"; fi

keys=$($CLI key list)
echo "$keys"| tr -d '\r' | awk '{ if ($2 != "username") printf "%s=%s\n",$2,$4 }' > testaddr.sh

