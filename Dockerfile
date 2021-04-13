FROM alpine:3.12

RUN addgroup -S ory; \
    adduser -S ory -G ory -D -u 10000 -h /home/ory -s /bin/nologin; \
    chown -R ory:ory /home/ory

RUN apk add -U --no-cache ca-certificates

COPY kratos /usr/bin/kratos

# Exposing the ory home directory to simplify passing in Kratos configuration (e.g. if the file $HOME/.kratos.yaml
# exists, it will be automatically used as the configuration file).
VOLUME /home/ory

# Declare the standard ports used by Kratos (4433 for public service endpoint, 4434 for admin service endpoint)
EXPOSE 4433 4434

USER 10000

ENTRYPOINT ["kratos"]
CMD ["serve"]
