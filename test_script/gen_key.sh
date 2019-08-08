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

# regenerate genesis key
amocli key remove tgenesis
amocli key import --username=tgenesis --encrypt=false "$GENESISPRIVKEY"

# regenerate validator, delegator keys
for ((i=1; i<=NODENUM; i++))
do
    amocli key remove tval$i
    amocli key generate tval$i --encrypt=false

    amocli key remove tdel$i
    amocli key generate tdel$i --encrypt=false
done

# regenerate user keys
amocli key remove tu1
amocli key generate tu1 --encrypt=false
amocli key remove tu2
amocli key generate tu2 --encrypt=false
