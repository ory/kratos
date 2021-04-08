SHELL=/bin/bash -o pipefail

#  EXECUTABLES = docker-compose docker node npm go
#  K := $(foreach exec,$(EXECUTABLES),\
#          $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE := on
export PATH := .bin:${PATH}
export PWD := $(shell pwd)

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/ory/x/tools/listx \
				  github.com/golang/mock/mockgen \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  golang.org/x/tools/cmd/goimports \
				  github.com/mikefarah/yq \
				  github.com/mattn/goveralls

define make-go-dependency
  # go install is responsible for not re-building when the code hasn't changed
  .bin/$(notdir $1): go.mod go.sum Makefile
		GOBIN=$(PWD)/.bin/ go install $1
endef
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))
$(call make-lint-dependency)

.bin/clidoc:
		go build -o .bin/clidoc ./cmd/clidoc/.

docs/cli: .bin/clidoc
		clidoc .

.bin/traefik:
		https://github.com/containous/traefik/releases/download/v2.3.0-rc4/traefik_v2.3.0-rc4_linux_amd64.tar.gz \
			tar -zxvf traefik_${traefik_version}_linux_${arch}.tar.gz

.bin/cli: go.mod go.sum Makefile
		go build -o .bin/cli -tags sqlite github.com/ory/cli

node_modules: package.json Makefile
		npm ci

docs/node_modules: docs/package.json
		cd docs; npm ci

.bin/golangci-lint: Makefile
		bash <(curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh) -d -b .bin v1.28.3

.bin/hydra: Makefile
		bash <(curl https://raw.githubusercontent.com/ory/hydra/master/install.sh) -d -b .bin v1.9.0-alpha.1

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

.PHONY: mocks
mocks: .bin/mockgen
		mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

.PHONY: install
install:
		GO111MODULE=on go install -tags sqlite .

.PHONY: test-resetdb
test-resetdb:
		script/testenv.sh

.PHONY: test
test:
		go test -p 1 -tags sqlite -count=1 -failfast ./...

.PHONY: test-coverage
test-coverage: .bin/go-acc .bin/goveralls
		go-acc -o coverage.txt ./... -- -v -failfast -timeout=20m -tags sqlite
		test -z "$CIRCLE_PR_NUMBER" && goveralls -service=circle-ci -coverprofile=coverage.txt -repotoken=$COVERALLS_REPO_TOKEN || echo "forks are not allowed to push to coveralls"

# Generates the SDK
.PHONY: sdk
sdk: .bin/swagger .bin/cli
		swagger generate spec -m -o spec/api.json -x internal/httpclient
		cli dev swagger sanitize ./spec/api.json
		swagger validate ./spec/api.json
		swagger flatten --with-flatten=remove-unused -o ./spec/api.json ./spec/api.json
		swagger validate ./spec/api.json
		rm -rf internal/httpclient/models/* internal/httpclient/clients/*
		mkdir -p internal/httpclient/
		swagger generate client -f ./spec/api.json -t internal/httpclient/ -A Ory_Kratos
		make format

.PHONY: quickstart
quickstart:
		docker pull oryd/kratos:latest-sqlite
		docker pull oryd/kratos-selfservice-ui-node:latest
		docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

.PHONY: quickstart-dev
quickstart-dev:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .
		docker-compose -f quickstart.yml -f quickstart-standalone.yml -f quickstart-latest.yml up --build --force-recreate

# Formats the code
.PHONY: format
format: .bin/goimports docs/node_modules node_modules
		goimports -w -local github.com/ory .
		cd docs; npm run format
		npm run format

# Runs tests in short mode, without database adapters
.PHONY: docker
docker:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .

# Runs the documentation tests
.PHONY: test-docs
test-docs: node_modules
		npm run text-run

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

.PHONY: migrations-render-replace
migrations-render-replace: .bin/cli
		cli dev pop migration render -r persistence/sql/migrations/templates persistence/sql/migrations/sql

.PHONY: migratest-refresh
migratest-refresh:
		cd persistence/sql/migratest; go test -tags sqlite,refresh -short .
