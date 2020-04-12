SHELL=/bin/bash -o pipefail

EXECUTABLES = docker-compose docker node npm go
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE := on
export PATH := .bin:${PATH}

deps:
ifneq ("v0",$(shell cat .bin/.lock))
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b .bin/ v1.24.0
		go build -o .bin/go-acc github.com/ory/go-acc
		go build -o .bin/goreturns github.com/sqs/goreturns
		go build -o .bin/listx github.com/ory/x/tools/listx
		go build -o .bin/mockgen github.com/golang/mock/mockgen
		go build -o .bin/swagger github.com/go-swagger/go-swagger/cmd/swagger
		go build -o .bin/goimports golang.org/x/tools/cmd/goimports
		go build -o .bin/swagutil github.com/ory/sdk/swagutil
		go build -o .bin/packr2 github.com/gobuffalo/packr/v2/packr2
		echo "v0" > .bin/.lock
endif

.PHONY: docs
docs:
		cd docs; npm i; npm run build

.PHONY: lint
lint: deps
		which golangci-lint
		GO111MODULE=on golangci-lint run -v ./...

.PHONY: cover
cover:
		go test ./... -coverprofile=cover.out
		go tool cover -func=cover.out

.PHONE: mocks
mocks: deps
		mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

.PHONY: install
install: deps
		packr2
		GO111MODULE=on go install .
		packr2 clean

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
sdk: deps
		$$(go env GOPATH)/bin/swagger generate spec -m -o .schema/api.swagger.json -x internal/httpclient
		$$(go env GOPATH)/bin/swagutil sanitize ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger validate ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger flatten --with-flatten=remove-unused -o ./.schema/api.swagger.json ./.schema/api.swagger.json
		$$(go env GOPATH)/bin/swagger validate ./.schema/api.swagger.json
		rm -rf internal/httpclient
		mkdir -p internal/httpclient
		$$(go env GOPATH)/bin/swagger generate client -f ./.schema/api.swagger.json -t internal/httpclient -A Ory_Kratos
		make format

.PHONY: quickstart
quickstart:
		docker pull oryd/kratos:latest-sqlite
		docker pull oryd/kratos-selfservice-ui-node:latest
		docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

.PHONY: quickstart-dev
quickstart-dev:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .
		docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

# Formats the code
.PHONY: format
format: deps
		$$(go env GOPATH)/bin/goreturns -w -local github.com/ory $$($$(go env GOPATH)/bin/listx .)

# Runs tests in short mode, without database adapters
.PHONY: docker
docker:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest .

.PHONY: test-e2e
test-e2e: