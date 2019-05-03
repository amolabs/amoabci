#!/bin/bash

. $(dirname $0)/env.sh

echo "stake of t0:" $(amocli query stake $t0)
echo "stake of t1:" $(amocli query stake $t1)
echo "stake of t2:" $(amocli query stake $t2)
echo "delegate of d0:" $(amocli query delegate $d0)
echo "delegate of d1:" $(amocli query delegate $d1)
echo "delegate of d2:" $(amocli query delegate $d2)


