#!/bin/bash

DATAROOT=$1
NODENUM=$2

# set basic environments
sed -e s#__dataroot__#$DATAROOT# -i.tmp docker-compose.yml

rm -rf $DATAROOT

mkdir -p $DATAROOT/seed/amo/config/
mkdir -p $DATAROOT/seed/amo/data/

for ((i=1; i<=NODENUM; i++))
do
    mkdir -p $DATAROOT/val$i/amo/config/
    mkdir -p $DATAROOT/val$i/amo/data/
done

