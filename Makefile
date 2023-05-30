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
export IMAGE_TAG 					:= $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/golang/mock/mockgen \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  golang.org/x/tools/cmd/goimports \
				  github.com/mattn/goveralls \
				  github.com/cortesi/modd/cmd/modd \
				  github.com/mailhog/MailHog

define make-go-dependency
  # go install is responsible for not re-building when the code hasn't changed
  .bin/$(notdir $1): go.mod go.sum
		GOBIN=$(PWD)/.bin/ go install $1
endef
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))
$(call make-lint-dependency)

.bin/clidoc:
	echo "deprecated usage, use docs/cli instead"
	go build -o .bin/clidoc ./cmd/clidoc/.

.PHONY: .bin/yq
.bin/yq:
	go build -o .bin/yq github.com/mikefarah/yq/v4

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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -d -b .bin v1.52.2

.bin/hydra: Makefile
	bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -d -b .bin hydra v2.0.2

.bin/ory: Makefile
	curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.2.2
	touch -a -m .bin/ory

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

test-short:
	go test -tags sqlite -count=1 -failfast -short ./...

.PHONY: test-coverage
test-coverage: .bin/go-acc .bin/goveralls
	go-acc -o coverage.out ./... -- -v -failfast -timeout=20m -tags sqlite

# Generates the SDK
.PHONY: sdk
sdk: .bin/swagger .bin/ory node_modules
	swagger generate spec -m -o spec/swagger.json \
		-c github.com/ory/kratos \
		-c github.com/ory/x/healthx \
		-c github.com/ory/x/openapix
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
		-t .schema/openapi/templates/go \
		-c .schema/openapi/gen.go.yml

	(cd internal/httpclient; rm -rf go.mod go.sum test api docs)

	rm -rf internal/httpclient-central
	mkdir -p internal/httpclient-central/
	npm run openapi-generator-cli -- generate -i "spec/api.json" \
		-g go \
		-o "internal/client-go" \
		--git-user-id ory \
		--git-repo-id client-go \
		--git-host github.com \
		-t .schema/openapi/templates/go \
		-c .schema/openapi/gen.go.yml

	(cd internal/client-go; go mod edit -module github.com/ory/client-go go.mod; rm -rf test api docs)

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
format: .bin/goimports .bin/ory node_modules
	.bin/ory dev headers copyright --exclude=internal/httpclient --exclude=internal/client-go --exclude test/e2e/proxy/node_modules --exclude test/e2e/node_modules --exclude node_modules
	goimports -w -local github.com/ory .
	npm exec -- prettier --write 'test/e2e/**/*{.ts,.js}'
	npm exec -- prettier --write '.github'

# Build local docker image
.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 docker build -f .docker/Dockerfile-build --build-arg=COMMIT=$(VCS_REF) --build-arg=BUILD_DATE=$(BUILD_DATE) -t oryd/kratos:${IMAGE_TAG} .

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

.PHONY: test-e2e-playwright
test-e2e-playwright: node_modules test-resetdb
	source script/test-envs.sh
	test/e2e/run.sh --only-setup
	(cd test/e2e; DB=memory npm run playwright)

.PHONY: migrations-sync
migrations-sync: .bin/ory
	ory dev pop migration sync persistence/sql/migrations/templates persistence/sql/migratest/testdata
	script/add-down-migrations.sh

.PHONY: test-update-snapshots
test-update-snapshots:
	UPDATE_SNAPSHOTS=true go test -p 4 -tags sqlite -short ./...

.PHONY: post-release
post-release: .bin/yq
	cat quickstart.yml | yq '.services.kratos.image = "oryd/kratos:'$$DOCKER_TAG'"' | sponge quickstart.yml
	cat quickstart.yml | yq '.services.kratos-migrate.image = "oryd/kratos:'$$DOCKER_TAG'"' | sponge quickstart.yml
	cat quickstart.yml | yq '.services.kratos-selfservice-ui-node.image = "oryd/kratos-selfservice-ui-node:'$$DOCKER_TAG'"' | sponge quickstart.yml

licenses: .bin/licenses node_modules  # checks open-source licenses
	.bin/licenses

.bin/licenses: Makefile
	curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

node_modules: package-lock.json
	npm ci
	touch node_modules
