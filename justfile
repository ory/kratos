# Ory Kratos justfile
# Install just: https://github.com/casey/just

# Set shell
set shell := ["bash", "-o", "pipefail", "-c"]

# Environment variables
export GO111MODULE := "on"
export PATH := env_var("PWD") / ".bin:" + env_var("PATH")
pwd := `pwd`
build_date := `date -u +"%Y-%m-%dT%H:%M:%SZ"`
vcs_ref := `git rev-parse HEAD`
image_tag := env_var_or_default("IMAGE_TAG", "latest")
quickstart_options := env_var_or_default("QUICKSTART_OPTIONS", "")

# Default recipe to display help information
default:
    @just --list

# Install Go dependencies
install-go-deps:
    GOBIN={{pwd}}/.bin/ go install github.com/ory/go-acc
    GOBIN={{pwd}}/.bin/ go install github.com/golang/mock/mockgen
    GOBIN={{pwd}}/.bin/ go install github.com/go-swagger/go-swagger/cmd/swagger
    GOBIN={{pwd}}/.bin/ go install golang.org/x/tools/cmd/goimports
    GOBIN={{pwd}}/.bin/ go install github.com/mattn/goveralls
    GOBIN={{pwd}}/.bin/ go install github.com/cortesi/modd/cmd/modd
    GOBIN={{pwd}}/.bin/ go install github.com/mailhog/MailHog

# Install clidoc (deprecated)
install-clidoc:
    @echo "deprecated usage, use docs/cli instead"
    go build -o .bin/clidoc ./cmd/clidoc/.

# Install yq
install-yq:
    GOBIN={{pwd}}/.bin go install github.com/mikefarah/yq/v4@v4.44.3

# Generate CLI documentation
docs-cli:
    go run ./cmd/clidoc/. .

# Preview API documentation
docs-api:
    npx @redocly/openapi-cli preview-docs spec/api.json

# Preview Swagger documentation
docs-swagger:
    npx @redocly/openapi-cli preview-docs spec/swagger.json

# Install golangci-lint
install-golangci-lint:
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -d -b .bin v1.61.0

# Install Hydra
install-hydra:
    bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -d -b .bin hydra v2.2.0-rc.3

# Install Ory CLI
install-ory:
    curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.2.2
    touch -a -m .bin/ory

# Install buf
install-buf:
    curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.39.0/buf-$(uname -s)-$(uname -m).tar.gz" | tar -xvzf - -C ".bin/" --strip-components=2 buf/bin/buf buf/bin/protoc-gen-buf-breaking buf/bin/protoc-gen-buf-lint
    touch -a -m .bin/buf

# Install licenses checker
install-licenses:
    curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

# Install npm dependencies
install-npm:
    npm ci
    npm audit fix --force
    touch node_modules
    cd test/e2e && npm ci && npm audit fix --force
    touch test/e2e/node_modules

# Install all dependencies
install-all: install-go-deps install-yq install-golangci-lint install-ory install-buf install-npm

# Run linters
lint: install-golangci-lint install-buf
    .bin/golangci-lint run -v --timeout 10m ./...
    .bin/buf lint

# Generate mocks
mocks:
    @mkdir -p .bin
    @if [ ! -f .bin/mockgen ]; then GOBIN={{pwd}}/.bin/ go install github.com/golang/mock/mockgen; fi
    mockgen -mock_names Manager=MockLoginExecutorDependencies -package internal -destination internal/hook_login_executor_dependencies.go github.com/ory/kratos/selfservice loginExecutorDependencies

# Generate protobuf code
proto: install-buf
    .bin/buf generate
    .bin/goimports -w gen/

# Install kratos binary
install:
    go install -tags sqlite .

# Reset test databases
test-resetdb:
    script/testenv.sh

# Run all tests
test:
    go test -p 1 -tags sqlite -count=1 -failfast ./...

# Run short tests (skip long-running tests)
test-short:
    go test -tags sqlite -count=1 -failfast -short ./...

# Run tests with coverage
test-coverage:
    @mkdir -p .bin
    @if [ ! -f .bin/go-acc ]; then GOBIN={{pwd}}/.bin/ go install github.com/ory/go-acc; fi
    @if [ ! -f .bin/goveralls ]; then GOBIN={{pwd}}/.bin/ go install github.com/mattn/goveralls; fi
    go-acc -o coverage.out ./... -- -failfast -timeout=20m -tags sqlite,json1

# Run tests with coverage (next generation)
test-coverage-next:
    @mkdir -p .bin coverage
    @if [ ! -f .bin/go-acc ]; then GOBIN={{pwd}}/.bin/ go install github.com/ory/go-acc; fi
    @if [ ! -f .bin/goveralls ]; then GOBIN={{pwd}}/.bin/ go install github.com/mattn/goveralls; fi
    go test -short -failfast -timeout=20m -tags sqlite,json1 -cover ./... --args test.gocoverdir="{{pwd}}/coverage"
    go tool covdata percent -i=coverage
    go tool covdata textfmt -i=./coverage -o coverage.new.out

# Run end-to-end tests
test-e2e: install-npm test-resetdb kratos-config-e2e
    source script/test-envs.sh
    test/e2e/run.sh sqlite
    test/e2e/run.sh postgres
    test/e2e/run.sh cockroach
    test/e2e/run.sh mysql

# Run Playwright end-to-end tests
test-e2e-playwright: install-npm test-resetdb kratos-config-e2e
    source script/test-envs.sh
    test/e2e/run.sh --only-setup
    cd test/e2e && DB=memory npm run playwright

# Refresh test snapshots
test-refresh:
    UPDATE_SNAPSHOTS=true go test -tags sqlite,json1,refresh -short ./...

# Generate Kratos config for e2e tests
kratos-config-e2e:
    sh ./test/e2e/render-kratos-config.sh

# Generate SDK
sdk: install-npm
    @mkdir -p .bin
    @if [ ! -f .bin/swagger ]; then GOBIN={{pwd}}/.bin/ go install github.com/go-swagger/go-swagger/cmd/swagger; fi
    @if [ ! -f .bin/ory ]; then just install-ory; fi
    swagger generate spec -m -o spec/swagger.json \
        -c github.com/ory/kratos \
        -c github.com/ory/x/healthx \
        -c github.com/ory/x/crdbx \
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
        --api-name-suffix "API" \
        -t .schema/openapi/templates/go \
        -c .schema/openapi/gen.go.yml
    cd internal/httpclient && rm -rf go.mod go.sum test api docs
    rm -rf internal/client-go
    mkdir -p internal/client-go/
    npm run openapi-generator-cli -- generate -i "spec/api.json" \
        -g go \
        -o "internal/client-go" \
        --git-user-id ory \
        --git-repo-id client-go \
        --git-host github.com \
        --api-name-suffix "API" \
        -t .schema/openapi/templates/go \
        -c .schema/openapi/gen.go.yml
    cd internal/client-go && go mod edit -module github.com/ory/client-go go.mod && rm -rf test api docs
    just format

# Format code
format: install-npm install-buf
    @if [ ! -f .bin/goimports ]; then GOBIN={{pwd}}/.bin/ go install golang.org/x/tools/cmd/goimports; fi
    @if [ ! -f .bin/ory ]; then just install-ory; fi
    .bin/ory dev headers copyright --exclude=gen --exclude=internal/httpclient --exclude=internal/client-go --exclude test/e2e/proxy/node_modules --exclude test/e2e/node_modules --exclude node_modules
    goimports -w -local github.com/ory .
    npm exec -- prettier --write 'test/e2e/**/*{.ts,.js}'
    npm exec -- prettier --write '.github'
    .bin/buf format --write

# Build Docker image
docker:
    DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 docker build -f .docker/Dockerfile-build --build-arg=COMMIT={{vcs_ref}} --build-arg=BUILD_DATE={{build_date}} -t oryd/kratos:{{image_tag}} .

# Run quickstart with Docker
quickstart:
    docker pull oryd/kratos:latest
    docker pull oryd/kratos-selfservice-ui-node:latest
    docker-compose -f quickstart.yml -f quickstart-standalone.yml up --build --force-recreate

# Run quickstart in dev mode
quickstart-dev:
    docker build -f .docker/Dockerfile-build -t oryd/kratos:latest .
    docker-compose -f quickstart.yml -f quickstart-standalone.yml -f quickstart-latest.yml {{quickstart_options}} up --build --force-recreate

# Update AUTHORS file
authors:
    curl https://raw.githubusercontent.com/ory/ci/master/authors/authors.sh | env PRODUCT="Ory Kratos" bash

# Sync database migrations
migrations-sync: install-ory
    ory dev pop migration sync persistence/sql/migrations/templates persistence/sql/migratest/testdata
    script/add-down-migrations.sh

# Post-release updates
post-release: install-yq
    #!/usr/bin/env bash
    cat quickstart.yml | yq '.services.kratos.image = "oryd/kratos:'$DOCKER_TAG'"' | sponge quickstart.yml
    cat quickstart.yml | yq '.services.kratos-migrate.image = "oryd/kratos:'$DOCKER_TAG'"' | sponge quickstart.yml
    cat quickstart.yml | yq '.services.kratos-selfservice-ui-node.image = "oryd/kratos-selfservice-ui-node:'$DOCKER_TAG'"' | sponge quickstart.yml

# Check open-source licenses
licenses: install-licenses install-npm
    .bin/licenses

# Clean build artifacts
clean:
    rm -rf .bin
    rm -rf node_modules
    rm -rf test/e2e/node_modules
    rm -rf coverage
    rm -f coverage.out coverage.new.out

# Run a specific package test
test-pkg package:
    go test -tags sqlite -v {{package}}

# Run a specific test function
test-func package func:
    go test -tags sqlite -run {{func}} {{package}}

# Build the project
build:
    go build -tags sqlite -o .bin/kratos .

# Run Kratos locally
run *ARGS:
    go run -tags sqlite . {{ARGS}}

# Show version info
version:
    @echo "Build Date: {{build_date}}"
    @echo "VCS Ref: {{vcs_ref}}"
    @echo "Image Tag: {{image_tag}}"

