SHELL=/usr/bin/env bash -o pipefail

#  EXECUTABLES = docker-compose docker node npm go
#  K := $(foreach exec,$(EXECUTABLES),\
#          $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

export GO111MODULE        := on
export PATH               := .bin:${PATH}
export PWD                := $(shell pwd)
export BUILD_DATE         := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
export VCS_REF            := $(shell git rev-parse HEAD)
export QUICKSTART_OPTIONS ?=
export IMAGE_TAG 					:= $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)

.bin/clidoc:
	echo "deprecated usage, use docs/cli instead"
	go build -o .bin/clidoc ./cmd/clidoc/.

.PHONY: docs/cli
docs/cli:
	go run ./cmd/clidoc/. .

.PHONY: docs/api
docs/api:
	npx @redocly/openapi-cli preview-docs spec/api.json

.PHONY: docs/swagger
docs/swagger:
	npx @redocly/openapi-cli preview-docs spec/swagger.json

.bin/golangci-lint: Makefile
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -d -b .bin v2.4.0

.bin/hydra: Makefile
	bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -d -b .bin hydra v2.2.0-rc.3

.bin/ory: Makefile
	curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.2.2
	touch -a -m .bin/ory

.bin/buf: Makefile
	curl -sSL \
	"https://github.com/bufbuild/buf/releases/download/v1.39.0/buf-$(shell uname -s)-$(shell uname -m).tar.gz" | \
	tar -xvzf - -C ".bin/" --strip-components=2 buf/bin/buf buf/bin/protoc-gen-buf-breaking buf/bin/protoc-gen-buf-lint
	touch -a -m .bin/buf

.PHONY: lint
lint: .bin/golangci-lint .bin/buf
	.bin/golangci-lint run -v --timeout 10m ./...
	.bin/buf lint

.PHONY: mocks
mocks:
	go tool mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

.PHONY: proto
proto: gen/oidc/v1/state.pb.go

gen/oidc/v1/state.pb.go: proto/oidc/v1/state.proto buf.yaml buf.gen.yaml .bin/buf
	.bin/buf generate
	go tool goimports -w gen/

.PHONY: install
install:
	go install -tags sqlite .

.PHONY: test-resetdb
test-resetdb:
	script/testenv.sh

.PHONY: test
test:
	docker pull oryd/hydra:v2.2.0-rc.3
	go test -p 1 -tags sqlite -count=1 -failfast ./...

test-short:
	go test -tags sqlite -count=1 -failfast -short ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out -failfast -timeout=20m -tags sqlite ./...

.PHONY: test-coverage-next
test-coverage-next:
	go test -short -failfast -timeout=20m -tags sqlite -cover ./... --args test.gocoverdir="$$PWD/coverage"
	go tool covdata percent -i=coverage
	go tool covdata textfmt -i=./coverage -o coverage.new.out

# Generates the SDK
.PHONY: sdk
sdk: .bin/ory node_modules
	go tool swagger generate spec -m -o spec/swagger.json \
		-c github.com/ory/kratos \
		-c github.com/ory/x/healthx \
		-c github.com/ory/x/crdbx \
		-c github.com/ory/x/openapix
	ory dev swagger sanitize ./spec/swagger.json
	go tool swagger validate ./spec/swagger.json
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
			-p file://.schema/openapi/patches/courier.yaml \
			-p file://.schema/openapi/patches/generic_error.yaml \
			-p file://.schema/openapi/patches/nulls.yaml \
			-p file://.schema/openapi/patches/common.yaml \
			spec/swagger.json spec/api.json

	rm -rf internal/httpclient
	mkdir -p internal/httpclient/
	npm run openapi-generator-cli -- generate -i "spec/api.json" \
		-g go \
		-o "internal/httpclient" \
		--git-user-id ory \
		--git-repo-id client-go \
		--git-host github.com \
		--api-name-suffix "API" \
		-c .schema/openapi/gen.go.yml

	(cd internal/httpclient; rm -rf go.mod go.sum test api docs)

	rm -rf internal/client-go
	mkdir -p internal/client-go/
	npm run openapi-generator-cli -- generate -i "spec/api.json" \
		-g go \
		-o "internal/client-go" \
		--git-user-id ory \
		--git-repo-id client-go \
		--git-host github.com \
		--api-name-suffix "API" \
		-c .schema/openapi/gen.go.yml

	(cd internal/client-go; go mod edit -module github.com/ory/client-go go.mod; rm -rf test api docs; go mod tidy)

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

authors:  # updates the AUTHORS file
	curl https://raw.githubusercontent.com/ory/ci/master/authors/authors.sh | env PRODUCT="Ory Kratos" bash

# Formats the code
.PHONY: format
format: .bin/ory node_modules .bin/buf
	.bin/ory dev headers copyright --exclude=gen --exclude=internal/httpclient --exclude=internal/client-go --exclude test/e2e/proxy/node_modules --exclude test/e2e/node_modules --exclude node_modules --exclude=oryx
	go tool goimports -w -local github.com/ory .
	npm exec -- prettier --write 'test/e2e/**/*{.ts,.js}'
	npm exec -- prettier --write '.github'
	.bin/buf format --write

# Build local docker image
.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 docker build -f .docker/Dockerfile-build --build-arg=COMMIT=$(VCS_REF) --build-arg=BUILD_DATE=$(BUILD_DATE) -t oryd/kratos:${IMAGE_TAG} .

.PHONY: test-e2e
test-e2e: node_modules test-resetdb kratos-config-e2e
	source script/test-envs.sh
	test/e2e/run.sh sqlite
	test/e2e/run.sh postgres
	test/e2e/run.sh cockroach
	test/e2e/run.sh mysql

.PHONY: test-e2e-playwright
test-e2e-playwright: node_modules test-resetdb kratos-config-e2e
	source script/test-envs.sh
	test/e2e/run.sh --only-setup
	(cd test/e2e; DB=memory npm run playwright)

.PHONY: test-refresh
test-refresh:
	UPDATE_SNAPSHOTS=true go test -tags sqlite,json1,refresh -short ./...

.PHONY: pre-release
pre-release:
	go tool yq '.services.kratos.image = "oryd/kratos:'$$DOCKER_TAG'"' -i quickstart.yml
	go tool yq '.services.kratos-migrate.image = "oryd/kratos:'$$DOCKER_TAG'"' -i quickstart.yml
	go tool yq '.services.kratos-selfservice-ui-node.image = "oryd/kratos-selfservice-ui-node:'$$DOCKER_TAG'"' -i quickstart.yml

.PHONY: post-release
post-release:
	echo "nothing to do"

licenses: .bin/licenses node_modules  # checks open-source licenses
	.bin/licenses

.bin/licenses: Makefile
	curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

node_modules: package-lock.json
	npm ci
	touch node_modules

.PHONY: kratos-config-e2e
kratos-config-e2e:
	sh ./test/e2e/render-kratos-config.sh
