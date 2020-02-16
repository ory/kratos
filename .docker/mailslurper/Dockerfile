FROM golang:1.13-alpine as builder

RUN apk --no-cache add git libc-dev gcc
RUN go get github.com/mjibson/esc

WORKDIR /go/src/github.com/mailslurper/mailslurper
RUN git clone https://github.com/mailslurper/mailslurper.git .

ADD . cmd/mailslurper

RUN cd cmd/mailslurper && \
    go get -v && \
    go generate && \
    go install -v

ADD . .

ENTRYPOINT mailslurper
