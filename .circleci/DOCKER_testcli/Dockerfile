FROM alpine:3.11

COPY amocli /usr/bin/

WORKDIR /

RUN apk add --no-cache bash bc curl

ENTRYPOINT [ "/bin/sh", "-c" ]

# to make the container run persistently
CMD [ "tail -f /dev/null" ]
