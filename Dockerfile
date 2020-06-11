# vim: set expandtab:

#### builder image

FROM golang:1.14-alpine3.12

ENV ROCKSDB_VER "6.10.1"
ENV LEVELDB_VER "1.20"

# tools
RUN apk add bash git make gcc g++

# libs
RUN apk add snappy snappy-dev

RUN mkdir /src
WORKDIR /src

# rocksdb 
RUN wget https://github.com/facebook/rocksdb/archive/v$ROCKSDB_VER.tar.gz 
RUN tar -xzf v$ROCKSDB_VER.tar.gz
RUN cp -a rocksdb-$ROCKSDB_VER/include/rocksdb /usr/include/
COPY contrib/rocksdb/librocksdb.so* /usr/lib/ 

# leveldb 
RUN wget https://github.com/google/leveldb/archive/v$LEVELDB_VER.tar.gz
RUN tar -xzf v$LEVELDB_VER.tar.gz
RUN cp -a leveldb-$LEVELDB_VER/include/leveldb /usr/include/
COPY contrib/leveldb/libleveldb.so* /usr/lib/ 

# amod
RUN mkdir -p amoabci
COPY Makefile go.mod go.sum amoabci/
COPY cmd amoabci/cmd
COPY amo amoabci/amo
COPY crypto amoabci/crypto
RUN make -C amoabci build_c

#### runner image

FROM alpine:3.12

# tools & libs
RUN apk add bash snappy

#COPY amod /usr/bin/
COPY --from=0 /usr/lib/librocksdb.so* /usr/lib/
COPY --from=0 /usr/lib/libleveldb.so* /usr/lib/
COPY --from=0 /usr/lib/libgcc_s.so* /usr/lib/
COPY --from=0 /usr/lib/libstdc++.so* /usr/lib/
COPY --from=0 /src/amoabci/amod /usr/bin/
COPY DOCKER/run_node.sh DOCKER/config/* /

ENV AMOHOME /amo
VOLUME [ $AMOHOME ]

WORKDIR /

EXPOSE 26656 26657

CMD ["/bin/sh", "/run_node.sh"]

