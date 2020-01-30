FROM alpine:3.11

RUN addgroup -S ory; \
    adduser -S ory -G ory -D -H -s /bin/nologin
RUN apk add -U --no-cache ca-certificates

COPY kratos /usr/bin/kratos

USER ory

ENTRYPOINT ["kratos"]
CMD ["serve"]
