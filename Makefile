SHELL=/bin/bash -o pipefail

all:
ifeq (, $(shell which golangci-lint))
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.2
endif

.PHONY: tools
tools:
		go install github.com/ory/go-acc github.com/ory/x/tools/listx github.com/go-swagger/go-swagger/cmd/swagger github.com/go-bindata/go-bindata/go-bindata github.com/sqs/goreturns github.com/ory/sdk/swagutil

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

.PHONY: cover
cover:
		go test ./... -coverprofile=cover.out
		go tool cover -func=cover.out

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

.PHONY: test-resetdb
test-resetdb:
		docker kill kratos_test_database_mysql || true
		docker kill kratos_test_database_postgres || true
		docker kill kratos_test_database_cockroach || true
		docker rm -f kratos_test_database_mysql || true
		docker rm -f kratos_test_database_postgres || true
		docker rm -f kratos_test_database_cockroach || true
		docker run --rm --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:5.7
		docker run --rm --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:9.6
		docker run --rm --name kratos_test_database_cockroach -p 3446:26257 -d cockroachdb/cockroach:v2.1.6 start --insecure

.PHONY: test
test: test-resetdb
		source scripts/test-envs.sh && go test -tags sqlite -count=1 ./...

# Generates the SDKs
.PHONY: sdk
sdk:
		$$(go env GOPATH)/bin/swagger generate spec -m -o .schema/api.swagger.json -x internal/httpclient
		$$(go env GOPATH)/bin/swagutil sanitize ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger validate ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger flatten --with-flatten=remove-unused -o ./.schema/api.swagger.json ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger validate ./.schema/api.swagger.json
		rm -rf internal/httpclient
		mkdir -p internal/httpclient
		$$(go env GOPATH)/bin/swagger generate client -f ./.schema/api.swagger.json -t internal/httpclient -A Ory_Kratos
		make format

quickstart:
		docker pull oryd/kratos:latest-sqlite
		docker pull oryd/kratos-selfservice-ui-node:latest
		docker-compose -f quickstart.yml up --build --force-recreate

quickstart-dev:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .
		docker-compose -f quickstart.yml up --build --force-recreate

# Formats the code
.PHONY: format
format:
		$$(go env GOPATH)/bin/goreturns -w -local github.com/ory $$($$(go env GOPATH)/bin/listx .)

# Runs tests in short mode, without database adapters
.PHONY: docker
docker:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest .
