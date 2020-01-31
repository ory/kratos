FROM alpine:3.11

RUN addgroup -S ory; \
    adduser -S ory -G ory -D  -h /home/ory -s /bin/nologin; \
    chown -R ory:ory /home/ory
RUN apk add -U --no-cache ca-certificates

WORKDIR /home/ory

# These are neccessary for mounting SQLITE volumes which require WRITE permission.
RUN mkdir -p /home/ory/sqlite; \
        chown ory:ory /home/ory/sqlite
VOLUME /home/ory/sqlite

COPY kratos /usr/bin/kratos

USER ory

ENTRYPOINT ["kratos"]
CMD ["serve"]
