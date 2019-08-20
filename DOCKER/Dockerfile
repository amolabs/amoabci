# vim: set expandtab:
FROM alpine:3.7

ENV AMOHOME /amo
ENV TMHOME /tendermint

VOLUME [ $AMOHOME ]
VOLUME [ $TMHOME ]

COPY tendermint amod /usr/bin/
COPY run_node.sh config/* /

WORKDIR /

EXPOSE 26656 26657

CMD ["/bin/sh", "/run_node.sh"]

