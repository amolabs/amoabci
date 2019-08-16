FROM alpine:3.7

COPY amocli /usr/bin/
COPY test_script/ /test_script/

WORKDIR /

RUN apk add --no-cache bash bc

ENTRYPOINT [ "/bin/sh", "-c" ]

# to make the container run persistently
CMD [ "tail -f /dev/null" ]