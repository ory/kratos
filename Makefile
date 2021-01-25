SHELL=/bin/bash -o pipefail

#  EXECUTABLES = docker-compose docker node npm go
#  K := $(foreach exec,$(EXECUTABLES),\
#          $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE := on
export PATH := .bin:${PATH}
export PWD := $(shell pwd)

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/ory/x/tools/listx \
				  github.com/markbates/pkger/cmd/pkger \
				  github.com/golang/mock/mockgen \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  golang.org/x/tools/cmd/goimports \
				  github.com/mikefarah/yq \
				  github.com/bufbuild/buf/cmd/buf \
					github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
					github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
					google.golang.org/protobuf/cmd/protoc-gen-go \
					google.golang.org/grpc/cmd/protoc-gen-go-grpc

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
install: pack
		GO111MODULE=on go install -tags sqlite .

.PHONY: test-resetdb
test-resetdb:
		script/testenv.sh

.PHONY: test
test:
		go test -p 1 -tags sqlite -count=1 -failfast ./...

# Generates the SDK
.PHONY: sdk
sdk: # .bin/swagger .bin/cli node_modules
		swagger generate spec -m -o .schema/api.swagger.json -x github.com/ory/kratos-client-go
		cli dev swagger sanitize ./.schema/api.swagger.json
		swagger validate ./.schema/api.swagger.json
		CIRCLE_PROJECT_USERNAME=ory CIRCLE_PROJECT_REPONAME=kratos \
				cli dev openapi migrate \
					-p https://raw.githubusercontent.com/ory/x/master/healthx/openapi/patch.yaml \
					-p file://.schema/openapi/patches/meta.yaml \
					-p file://.schema/openapi/patches/schema.yaml \
					.schema/api.swagger.json .schema/api.openapi.json

		npm run openapi-generator-cli -- generate -i ".schema/api.openapi.json" \
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
		docker pull oryd/kratos:latest-sqlite
		docker pull oryd/kratos-selfservice-ui-node:latest
		docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

.PHONY: quickstart-dev
quickstart-dev:
		docker build -f .docker/Dockerfile-build -t oryd/kratos:latest-sqlite .
		docker-compose -f quickstart.yml -f quickstart-standalone.yml -f quickstart-latest.yml up --build --force-recreate

# Formats the code
.PHONY: format
format: .bin/goimports
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

.PHONY: pack
pack: .bin/pkger
		pkger -exclude node_modules -exclude docs -exclude .git -exclude .github -exclude .bin -exclude test -exclude script -exclude contrib

.PHONY: buf-tools
buf-tools: .bin/buf .bin/protoc-gen-grpc-gateway .bin/protoc-gen-openapiv2 .bin/protoc-gen-go .bin/protoc-gen-go-grpc

.PHONY: buf-lint
buf-lint: buf-tools
		buf lint

.PHONY: buf-gen
buf-gen: buf-tools
		buf generate
