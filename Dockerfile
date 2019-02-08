FROM golang:alpine as builder

ENV PACKAGES make git libc-dev bash gcc linux-headers eudev-dev
ENV DIR /go/src/github.com/amolabs/amoabci

WORKDIR $DIR

COPY . .

RUN apk add --no-cache $PACKAGES \
	&& make tools \
	&& make vendor-deps \
	&& make build

FROM alpine:edge

RUN apk add --update ca-certificates

WORKDIR /root

COPY --from=builder /go/src/github.com/amolabs/amoabci/amod /usr/bin/amod

EXPOSE 26656 26657 26658

CMD ["amod"]