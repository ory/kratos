FROM alpine:3.11


RUN addgroup -S ory; \
    adduser -S ory -G ory -D -s /bin/nologin
RUN apk add -U --no-cache ca-certificates

WORKDIR /home/ory

COPY kratos /usr/bin/kratos

# These directories must exist for packr to work. Directory `/root/project` is the one where the goreleaser orb
# builds the binary and copies the source code to.
RUN mkdir -p contrib/sql/migrations docs
USER ory

ENTRYPOINT ["kratos"]
CMD ["serve"]
