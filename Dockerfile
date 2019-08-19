FROM alpine:3.9

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY hive /usr/bin/hive

USER 1000

ENTRYPOINT ["hive"]
CMD ["serve"]
