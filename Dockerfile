# vim: set expandtab:
FROM golang:alpine as builder

ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev
ENV DIR /go/src/github.com/amolabs/amoabci

WORKDIR $DIR

COPY Makefile Gopkg.toml Gopkg.lock $DIR/

RUN apk add --no-cache $PACKAGES \
    && make get_tools \
    && make get_vendor_deps

COPY main.go $DIR/
COPY amo $DIR/amo
RUN make TARGET=linux

FROM amolabs/tendermint-amo:latest

# tendermint base image uses /tendermint as a home directory
WORKDIR /tendermint

#RUN apk add --update ca-certificates

COPY --from=builder /go/src/github.com/amolabs/amoabci/amod /usr/bin/amod
COPY run.sh config/* ./

EXPOSE 26656 26657

# We need to override ENTRYPOINT from tendermint base image.
ENTRYPOINT ["/bin/sh"]
CMD ["/tendermint/run.sh"]
