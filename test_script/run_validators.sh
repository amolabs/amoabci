#!/bin/bash

NODENUM=$1

# val nodes: val2, val3, val4, val5, val6
for ((i=2; i<=NODENUM; i++))
do
    docker-compose up -d val$i
done
