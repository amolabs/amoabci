# vim: set expandtab:
FROM golang:alpine as builder

ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev
ENV DIR /go/src/github.com/amolabs/amoabci

WORKDIR $DIR

COPY Makefile Gopkg.toml Gopkg.lock $DIR/

RUN apk add --no-cache $PACKAGES \
    && make tools \
    && make vendor-deps

COPY main.go $DIR/
COPY amo $DIR/amo
RUN make TARGET=linux

FROM alpine:edge

RUN apk add --update ca-certificates

WORKDIR /root

COPY --from=builder /go/src/github.com/amolabs/amoabci/amod /usr/bin/amod
COPY run.sh /root
COPY config/* /root/

EXPOSE 26656 26657

# TODO: use ENTRYPOINT
#CMD ["amod"]
CMD ["/bin/sh", "/root/run.sh"]
