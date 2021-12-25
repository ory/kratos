SHELL=/bin/bash -o pipefail

#  EXECUTABLES = docker-compose docker node npm go
#  K := $(foreach exec,$(EXECUTABLES),\
#          $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE        := on
export PATH               := .bin:${PATH}
export PWD                := $(shell pwd)
export BUILD_DATE         := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
export VCS_REF            := $(shell git rev-parse HEAD)
export QUICKSTART_OPTIONS ?= ""

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/ory/x/tools/listx \
				  github.com/golang/mock/mockgen \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  golang.org/x/tools/cmd/goimports \
				  github.com/mikefarah/yq \
				  github.com/mattn/goveralls \
				  github.com/cortesi/modd/cmd/modd

define make-go-dependency
  # go install is responsible for not re-building when the code hasn't changed
  .bin/$(notdir $1): go.mod go.sum Makefile
		GOBIN=$(PWD)/.bin/ go install $1
endef
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))
$(call make-lint-dependency)

.bin/clidoc:
		echo "deprecated usage, use docs/cli instead"
		go build -o .bin/clidoc ./cmd/clidoc/.

.PHONY: docs/cli
docs/cli:
		go run ./cmd/clidoc/. .

.bin/ory: Makefile
		bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -d -b .bin ory v0.1.0
		touch -a -m .bin/ory

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
		golangci-lint run -v --timeout 10m ./...

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
		go-acc -o coverage.out ./... -- -v -failfast -timeout=20m -tags sqlite

# Generates the SDK
.PHONY: sdk
sdk: .bin/swagger .bin/ory node_modules
		swagger generate spec -m -o spec/swagger.json \
			-c github.com/ory/kratos \
			-c github.com/ory/x/healthx
		ory dev swagger sanitize ./spec/swagger.json
		swagger validate ./spec/swagger.json
		CIRCLE_PROJECT_USERNAME=ory CIRCLE_PROJECT_REPONAME=kratos \
				ory dev openapi migrate \
					--health-path-tags metadata \
					-p https://raw.githubusercontent.com/ory/x/master/healthx/openapi/patch.yaml \
					-p file://.schema/openapi/patches/meta.yaml \
					-p file://.schema/openapi/patches/schema.yaml \
					-p file://.schema/openapi/patches/selfservice.yaml \
					-p file://.schema/openapi/patches/security.yaml \
					-p file://.schema/openapi/patches/session.yaml \
					-p file://.schema/openapi/patches/identity.yaml \
					-p file://.schema/openapi/patches/generic_error.yaml \
					spec/swagger.json spec/api.json

		rm -rf internal/httpclient
		mkdir -p internal/httpclient/
		npm run openapi-generator-cli -- generate -i "spec/api.json" \
				-g go \
				-o "internal/httpclient" \
				--git-user-id ory \
				--git-repo-id kratos-client-go \
				--git-host github.com \
				-t .schema/openapi/templates/go \
				-c .schema/openapi/gen.go.yml

		make format

.PHONY: quickstart
quickstart:
		docker pull oryd/kratos:latest
		docker pull oryd/kratos-selfservice-ui-node:latest
		docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

.PHONY: quickstart-dev
quickstart-dev:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest .
		docker-compose -f quickstart.yml -f quickstart-standalone.yml -f quickstart-latest.yml $(QUICKSTART_OPTIONS) up --build --force-recreate

# Formats the code
.PHONY: format
format: .bin/goimports docs/node_modules node_modules
		goimports -w -local github.com/ory .
		cd docs; npm run format
		npm run format

# Build local docker image
.PHONY: docker
docker:
		DOCKER_BUILDKIT=1 docker build -f .docker/Dockerfile-build --build-arg=COMMIT=$(VCS_REF) --build-arg=BUILD_DATE=$(BUILD_DATE) -t oryd/kratos:latest .

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
migrations-sync: .bin/ory
		ory dev pop migration sync persistence/sql/migrations/templates persistence/sql/migratest/testdata

.PHONY: migrations-render
migrations-render: .bin/ory
		ory dev pop migration render persistence/sql/migrations/templates persistence/sql/migrations/sql

.PHONY: migrations-render-replace
migrations-render-replace: .bin/ory
		ory dev pop migration render -r persistence/sql/migrations/templates persistence/sql/migrations/sql

.PHONY: migratest-refresh
migratest-refresh:
		cd persistence/sql/migratest; UPDATE_SNAPSHOTS=true go test -p 1 -tags sqlite -short .

.PHONY: test-update-snapshots
test-update-snapshots:
		UPDATE_SNAPSHOTS=true go test -p 4 -tags sqlite -short ./...
