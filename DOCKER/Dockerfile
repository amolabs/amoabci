# vim: set expandtab:
FROM tendermint/tendermint:latest

# tendermint base image uses /tendermint as a home directory
WORKDIR /tendermint

#RUN apk add --update ca-certificates

COPY amod amocli /usr/bin/
COPY run.sh config/* ./
COPY --chown=tmuser:tmuser config/keys.json ./.amocli/keys/

EXPOSE 26656 26657

# We need to override ENTRYPOINT from tendermint base image.
ENTRYPOINT ["/bin/sh"]
CMD ["/tendermint/run.sh"]
