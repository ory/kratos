SHELL=/bin/bash -o pipefail

all:
ifeq (, $(shell which golangci-lint))
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1
endif

.PHONY: init
init:
		GO111MODULE=off go get -u \
			github.com/ory/x/tools/listx \
			github.com/sqs/goreturns \
			github.com/ory/go-acc \
			github.com/golang/mock/mockgen \
			github.com/go-swagger/go-swagger/cmd/swagger \
			golang.org/x/tools/cmd/goimports \
			github.com/go-openapi/errors \
			github.com/go-openapi/runtime \
			github.com/go-openapi/runtime/client \
			github.com/go-openapi/strfmt \
			github.com/golang/mock/...

.PHONY: lint
lint:
		golangci-lint run

.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)

.PHONY: cover
cover:
		go test ./... -coverprofile=cover.out
		go tool cover -func=cover.out

.PHONY: sdk
sdk:
		GO111MODULE=on go mod tidy
		GO111MODULE=on go mod vendor
		GO111MODULE=off swagger generate spec -x sdk/go/hive -m -o ./docs/api.swagger.json
		GO111MODULE=off swagger validate ./docs/api.swagger.json

		rm -rf ./sdk/go/hive/*
		GO111MODULE=off swagger generate client -f ./docs/api.swagger.json -t sdk/go/hive -A Ory_Hive

		cd sdk/go/hive; goreturns -w -i -local github.com/ory $$(listx .)

		rm -rf ./vendor

.PHONE: mocks
mocks:
		mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/hive/selfservice loginExecutorDependencies

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
		docker run --rm --name hydra_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=hydra -d postgres:9.6
