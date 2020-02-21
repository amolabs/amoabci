# vim: set expandtab:

#### builder image

FROM golang:1.13-alpine3.11

# tools
RUN apk add bash git make gcc g++

# libs
RUN apk add snappy

RUN mkdir /src
WORKDIR /src

# leveldb
RUN wget https://github.com/google/leveldb/archive/v1.20.tar.gz
RUN tar zxvf v1.20.tar.gz 
RUN cp -a leveldb-1.20/include/leveldb /usr/include/
COPY contrib/leveldb/bin/libleveldb.so* /usr/lib/ 

# amod
RUN mkdir -p amoabci
COPY Makefile go.mod go.sum amoabci/
COPY cmd amoabci/cmd
COPY amo amoabci/amo
COPY crypto amoabci/crypto
RUN make -C amoabci build_c

#### runner image

FROM alpine:3.11

# tools & libs
RUN apk add bash snappy

#COPY tendermint amod /usr/bin/
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

