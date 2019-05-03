#!/bin/bash

. $(dirname $0)/env.sh

echo "stake of t0:" $(amocli query stake $t0)
echo "stake of t1:" $(amocli query stake $t1)
echo "stake of t2:" $(amocli query stake $t2)
echo "stake of u0:" $(amocli query delegate $u0)
echo "stake of u1:" $(amocli query delegate $u1)
echo "stake of u2:" $(amocli query delegate $u2)


