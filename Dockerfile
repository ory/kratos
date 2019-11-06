FROM alpine:3.10

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY kratos /usr/bin/kratos

USER 1000

ENTRYPOINT ["kratos"]
CMD ["serve"]
