FROM golang:1.13.5-alpine3.10 AS build-env

# Set the Current Working Directory inside the container
WORKDIR /go/src/github.com/ory/kratos/

# # Install all dependancy 
RUN apk --no-cache add build-base git bzr mercurial gcc

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go mod download
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o kratos .

# Start secound stage build 
FROM alpine:3.10

RUN apk add -U --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the build image to the Working Directory inside the container
COPY --from=build-env /go/src/github.com/ory/kratos/kratos /usr/bin/kratos

USER 1000

ENTRYPOINT ["kratos"]
CMD ["serve"]
