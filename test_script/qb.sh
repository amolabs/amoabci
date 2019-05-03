#!/bin/bash

. $(dirname $0)/env.sh

echo "balance of t0:" $(amocli query balance $t0)
echo "balance of t1:" $(amocli query balance $t1)
echo "balance of t2:" $(amocli query balance $t2)
echo "balance of d0:" $(amocli query balance $d0)
echo "balance of d1:" $(amocli query balance $d1)
echo "balance of d2:" $(amocli query balance $d2)
echo "balance of u0:" $(amocli query balance $u0)
echo "balance of u1:" $(amocli query balance $u1)
echo "balance of u2:" $(amocli query balance $u2)

