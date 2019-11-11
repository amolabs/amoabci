#!/bin/bash

NAME="cleveldb_builder"
IMAGE="golang:1.12-alpine3.9"

CURRENT_PATH=$(pwd)/bin
ARTIFACT=/artifact

# run docker image
docker run -d -it \
	--name=$NAME \
	--volume $CURRENT_PATH:$ARTIFACT \
	--rm \
	$IMAGE

# tools
docker exec -it $NAME apk add bash git make gcc g++ snappy

# build
docker exec -it $NAME bash -c \
	"wget https://github.com/google/leveldb/archive/v1.20.tar.gz;
	tar zxvf v1.20.tar.gz && make -C leveldb-1.20;
	cp -a leveldb-1.20/out-shared/libleveldb.so* $ARTIFACT"

# stop docker container
docker stop $NAME
