FROM golang:1.12-alpine

ARG git_tag
ARG git_commit

RUN apk add --no-cache git build-base

WORKDIR /go/src/github.com/ory/hive-cloud/hive

ENV GO111MODULE=on

ADD ./go.mod ./go.mod
ADD ./go.sum ./go.sum

RUN go mod download

ADD . .

RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags "-s -X github.com/ory/hive-cloud/hive-api/cmd.BuildVersion=$git_tag -X github.com/ory/hive-cloud/hive-api/cmd.BuildTime=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/hive-cloud/hive-api/cmd.GitHash=$git_commit" \
    -a -installsuffix cgo -o \
    hive

FROM scratch

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/ory/hive-cloud/hive-api/hive-api /usr/bin/hive-api

ENTRYPOINT ["hive"]

CMD ["serve", "all"]
