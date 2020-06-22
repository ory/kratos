SHELL=/bin/bash -o pipefail

EXECUTABLES = docker-compose docker node npm go
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE := on
export PATH := .bin:${PATH}

.PHONY: deps
deps:
ifneq ("$(shell base64 Makefile))","$(shell cat .bin/.lock)")
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b .bin/ v1.24.0
		bash <(curl https://raw.githubusercontent.com/ory/hydra/master/install.sh) -b .bin/ v1.4.10
		go build -o .bin/go-acc github.com/ory/go-acc
		go build -o .bin/goreturns github.com/sqs/goreturns
		go build -o .bin/listx github.com/ory/x/tools/listx
		go build -o .bin/mockgen github.com/golang/mock/mockgen
		go build -o .bin/swagger github.com/go-swagger/go-swagger/cmd/swagger
		go build -o .bin/goimports golang.org/x/tools/cmd/goimports
		go build -o .bin/packr2 github.com/gobuffalo/packr/v2/packr2
		go build -o .bin/yq github.com/mikefarah/yq
		go build -o .bin/ory github.com/ory/cli
		npm ci
		echo "$$(base64 Makefile)" > .bin/.lock
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
		GO111MODULE=on go install -tags sqlite .
		packr2 clean

.PHONY: test-resetdb
test-resetdb:
		script/testenv.sh

.PHONY: test
test: test-resetdb
		source script/test-envs.sh && go test -p 1 -tags sqlite -count=1 -failfast ./...

# Generates the SDK
.PHONY: sdk
sdk: deps
		swagger generate spec -m -o .schema/api.swagger.json -x internal/httpclient
		ory dev swagger sanitize ./.schema/api.swagger.json
		swagger validate ./.schema/api.swagger.json
		swagger flatten --with-flatten=remove-unused -o ./.schema/api.swagger.json ./.schema/api.swagger.json
		swagger validate ./.schema/api.swagger.json
		rm -rf internal/httpclient
		mkdir -p internal/httpclient
		swagger generate client -f ./.schema/api.swagger.json -t internal/httpclient -A Ory_Kratos
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
		goreturns -w -local github.com/ory $$(listx .)
		npm run format

# Runs tests in short mode, without database adapters
.PHONY: docker
docker:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest .

.PHONY: test-e2e
test-e2e: test-resetdb
		source script/test-envs.sh
		test/e2e/run.sh sqlite
		test/e2e/run.sh postgres
		test/e2e/run.sh cockroach
		test/e2e/run.sh mysql

.PHONY: migrations-sync
migrations-sync:
		ory dev pop migration sync persistence/sql/migrations/templates persistence/sql/migratest/testdata

.PHONY: migrations-render
migrations-render:
		ory dev pop migration render persistence/sql/migrations/templates persistence/sql/migrations/sql
