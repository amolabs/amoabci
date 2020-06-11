# vim: set expandtab:

#### builder image

FROM golang:1.14-alpine3.12

# tools
RUN apk add bash git make gcc g++

# libs
RUN apk add snappy snappy-dev

RUN mkdir /src
WORKDIR /src

# rocksdb 
RUN wget https://github.com/facebook/rocksdb/archive/v6.10.1.tar.gz 
RUN tar -xzf v6.10.1.tar.gz
RUN cp -a rocksdb-6.10.1/include/rocksdb /usr/include/
COPY contrib/rocksdb/librocksdb.so* /usr/lib/ 

# leveldb 
RUN wget https://github.com/google/leveldb/archive/v1.20.tar.gz
RUN tar -xzf v1.20.tar.gz
RUN cp -a leveldb-1.20/include/leveldb /usr/include/
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
COPY --from=0 /usr/lib/librocksdb.so.6.10.1 /usr/lib/
RUN ln -sf /usr/lib/librocksdb.so.6.10.1 /usr/lib/librocksdb.so
RUN ln -sf /usr/lib/librocksdb.so.6.10.1 /usr/lib/librocksdb.so.6
RUN ln -sf /usr/lib/librocksdb.so.6.10.1 /usr/lib/librocksdb.so.6.10
COPY --from=0 /usr/lib/libleveldb.so.1.20 /usr/lib/
RUN ln -sf /usr/lib/libleveldb.so.1.20 /usr/lib/libleveldb.so.1
RUN ln -sf /usr/lib/libleveldb.so.1.20 /usr/lib/libleveldb.so
COPY --from=0 /usr/lib/libgcc_s.so* /usr/lib/
COPY --from=0 /usr/lib/libstdc++.so.6.0.28 /usr/lib/
RUN ln -sf /usr/lib/libstdc++.so.6.0.28 /usr/lib/libstdc++.so.6.0
RUN ln -sf /usr/lib/libstdc++.so.6.0.28 /usr/lib/libstdc++.so.6
RUN ln -sf /usr/lib/libstdc++.so.6.0.28 /usr/lib/libstdc++.so

COPY --from=0 /src/amoabci/amod /usr/bin/
COPY DOCKER/run_node.sh DOCKER/config/* /

ENV AMOHOME /amo
VOLUME [ $AMOHOME ]

WORKDIR /

EXPOSE 26656 26657

CMD ["/bin/sh", "/run_node.sh"]

