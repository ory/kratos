FROM alpine:3.10

RUN apk add -U --no-cache ca-certificates
COPY kratos /usr/bin/kratos

USER 1000

ENTRYPOINT ["kratos"]
CMD ["serve"]
