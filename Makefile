SHELL=/bin/bash -o pipefail

EXECUTABLES = docker-compose docker node npm go
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE := on
export PATH := .bin:${PATH}

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/sqs/goreturns \
				  github.com/ory/x/tools/listx \
				  github.com/golang/mock/mockgen \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  golang.org/x/tools/cmd/goimports \
				  github.com/ory/cli \
				  github.com/mikefarah/yq \
				  github.com/markbates/pkger/cmd/pkger \
				  github.com/gobuffalo/packr/v2/packr2

define make-go-dependency
  # go install is responsible for not re-building when the code hasn't changed
  .PHONY: .bin/$(notdir $1)
  .bin/$(notdir $1):
		GOBIN=$(PWD)/.bin/ go install $1
endef
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))
$(call make-lint-dependency)

node_modules: package.json
		npm ci

docs/node_modules: docs/package.json
		cd docs; npm ci

.bin/golangci-lint: Makefile
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b .bin v1.28.3

.bin/hydra: Makefile
		bash <(curl https://raw.githubusercontent.com/ory/hydra/master/install.sh) -b .bin v1.6.0

.PHONY: docs
docs: docs/node_modules
		cd docs; npm run build

.PHONY: lint
lint: .bin/golangci-lint
		golangci-lint run -v ./...

.PHONY: cover
cover:
		go test ./... -coverprofile=cover.out
		go tool cover -func=cover.out

.PHONE: mocks
mocks: .bin/mockgen
		mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

.PHONY: install
install: .bin/packr2
		packr2
		GO111MODULE=on go install -tags sqlite .
		packr2 clean

.PHONY: test-resetdb
test-resetdb:
		script/testenv.sh

.PHONY: test
test:
		go test -p 1 -tags sqlite -count=1 -failfast ./...

# Generates the SDK
.PHONY: sdk
sdk: .bin/swagger .bin/cli
		swagger generate spec -m -o .schema/api.swagger.json -x internal/httpclient
		cli dev swagger sanitize ./.schema/api.swagger.json
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
format: .bin/goreturns
		goreturns -w -local github.com/ory $$(listx .)
		npm run format

# Runs tests in short mode, without database adapters
.PHONY: docker
docker:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .

.PHONY: test-e2e
test-e2e: node_modules test-resetdb
		source script/test-envs.sh
		test/e2e/run.sh sqlite
		test/e2e/run.sh postgres
		test/e2e/run.sh cockroach
		test/e2e/run.sh mysql

.PHONY: migrations-sync
migrations-sync: .bin/cli
		cli dev pop migration sync persistence/sql/migrations/templates persistence/sql/migratest/testdata

.PHONY: migrations-render
migrations-render: .bin/cli
		cli dev pop migration render persistence/sql/migrations/templates persistence/sql/migrations/sql
