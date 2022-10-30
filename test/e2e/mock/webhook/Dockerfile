FROM golang:1.19-alpine AS build

WORKDIR /build

ADD . .

RUN cat go.*
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w -extldflags '-static'" -o ./app main.go

FROM alpine

WORKDIR /app

COPY --from=build /build/app ./mock

USER 65534:65534

EXPOSE 4459
ENTRYPOINT ["/app/mock"]
