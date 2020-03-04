# vim: set expandtab:

#### builder image

FROM golang:1.13-alpine3.11

# tools
RUN apk add bash git make gcc g++

# libs
RUN apk add snappy

RUN mkdir /src
WORKDIR /src

# amod
RUN mkdir -p amoabci
COPY Makefile go.mod go.sum amoabci/
COPY cmd amoabci/cmd
COPY amo amoabci/amo
COPY crypto amoabci/crypto
RUN make -C amoabci build

#### runner image

FROM alpine:3.11

# tools & libs
RUN apk add bash snappy

COPY --from=0 /src/amoabci/amod /usr/bin/
COPY DOCKER/run_node.sh DOCKER/config/* /

ENV AMOHOME /amo
VOLUME [ $AMOHOME ]

WORKDIR /

EXPOSE 26656 26657

CMD ["/bin/sh", "/run_node.sh"]

