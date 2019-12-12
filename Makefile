SHELL=/bin/bash -o pipefail

all:
ifeq (, $(shell which golangci-lint))
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1
endif

.PHONY: build
build:
		make sqlbin
		CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o kratos .

.PHONY: init
init:
		go install \
			github.com/sqs/goreturns \
			github.com/ory/x/tools/listx \
			github.com/ory/go-acc \
			github.com/golang/mock/mockgen \
			github.com/go-swagger/go-swagger/cmd/swagger \
			golang.org/x/tools/cmd/goimports

.PHONY: lint
lint:
		GO111MODULE=on golangci-lint run -v ./...

.PHONY: format
format:
		$$(go env GOPATH)/bin/goreturns -w -local github.com/ory $$($$(go env GOPATH)/bin/listx .)

.PHONY: cover
cover:
		go test ./... -coverprofile=cover.out
		go tool cover -func=cover.out

.PHONY: sdk
sdk:
		GO111MODULE=on go mod tidy
		GO111MODULE=on $$(go env GOPATH)/bin/swagger generate spec -x sdk/go/kratos -m -o ./docs/api.swagger.json
		GO111MODULE=on $$(go env GOPATH)/bin/swagger validate ./docs/api.swagger.json
		GO111MODULE=on go run ./contrib/swagutil sanitize ./docs/api.swagger.json

		rm -rf ./sdk/go/kratos/*
		GO111MODULE=on $$(go env GOPATH)/bin/swagger generate client --allow-template-override -f ./docs/api.swagger.json -t sdk/go/kratos -A Ory_Kratos

		cd sdk/go/kratos; goreturns -w -i -local github.com/ory $$(listx .)

.PHONE: mocks
mocks:
		mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

.PHONY: install
install:
		packr2 || (GO111MODULE=on go install github.com/gobuffalo/packr/v2/packr2 && packr2)
		GO111MODULE=on go install .
		packr2 clean

# Adds sql files to the binary using go-bindata
.PHONY: sqlbin
sqlbin:
		cd driver; go-bindata -o sql_migration_files.go -pkg driver ../contrib/sql/...

# Resets the test databases
.PHONY: resetdb
resetdb:
		docker kill hydra_test_database_postgres || true
		docker rm -f hydra_test_database_postgres || true
		docker run --rm --name hydra_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:9.6
