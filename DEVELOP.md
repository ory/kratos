# Development

This document explains how to develop Ory Kratos, run tests, and work with the tooling around it.

## Upgrading and changelog

Check [releases tab](https://github.com/ory/kratos/releases) for updates and changelogs when using the open source license.

## Command line documentation

To see available commands and flags, run:

```bash
kratos -h
# or
kratos help
```

## Contribution guidelines

We encourage all contributions. Before opening a pull request, read the
[contribution guidelines](./CONTRIBUTING.md).

## Prerequisites

You need Go 1.16+ and, for the test suites:

* Docker and Docker Compose
* `make`
* Node.js and npm

You can develop Ory Kratos on Windows, but most guides assume a Unix shell such as `bash` or `zsh`.

## Install from source

To install Kratos from source:

```make
make install
```

## Formatting code

Format all code using:

```make
make format
```

The continuous integration pipeline checks code formatting.

## Running tests

There are three types of tests:

* Short tests that do not require a SQL database
* Regular tests that require PostgreSQL, MySQL, and CockroachDB
* End to end tests that use real databases and a test browser

### Short tests

Short tests run quickly and use SQLite.

Run all short tests:

```bash
go test -short -tags sqlite ./...
```

Run short tests in a specific module:

```bash
cd client
go test -short -tags sqlite .
```

### Regular tests

Regular tests require a database setup.

The test suite can start databases using
[ory/dockertest](https://github.com/ory/dockertest). In practice, it is usually
easier and faster to use the Makefile targets.

Run the full test suite:

```make
make test
```

> Note: `make test` recreates the databases every time. This can be slow if you
> are iterating frequently on a specific test.

If you want to reuse databases across test runs, initialize them once:

```bash
make test-resetdb
export TEST_DATABASE_MYSQL='mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true'
export TEST_DATABASE_POSTGRESQL='postgres://postgres:secret@127.0.0.1:3445/kratos?sslmode=disable'
export TEST_DATABASE_COCKROACHDB='cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable'
```

Then you can run Go tests directly as often as needed:

```bash
go test -tags sqlite ./...

# or in a module:
cd client
go test -tags sqlite .
```

### Updating test fixtures

Some tests use snapshot fixtures.

Update snapshots for short tests:

```bash
make test-update-snapshots
```

Update all snapshots:

```bash
UPDATE_SNAPSHOTS=true go test -p 4 -tags sqlite ./...
```

You can run this from the repository root or from subdirectories.

### End-to-end tests

End to end tests are implemented with [Cypress](https://www.cypress.io).

> On ARM based Macs you may need to install Rosetta 2 to run Cypress.
> See the Cypress documentation:
> [https://www.cypress.io/blog/2021/01/20/running-cypress-on-the-apple-m1-silicon-arm-architecture-using-rosetta-2/](https://www.cypress.io/blog/2021/01/20/running-cypress-on-the-apple-m1-silicon-arm-architecture-using-rosetta-2/)

To install Rosetta 2:

```bash
softwareupdate --install-rosetta --agree-to-license
```

Run e2e tests in development mode:

```bash
./test/e2e/run.sh --dev sqlite
```

Run all e2e tests with databases:

```make
make test-e2e
```

For more options:

```bash
./test/e2e/run.sh
```

#### Run a single test

Add `.only` to the test you want to run, for example:

```ts
it.only('invalid remote recovery email template', () => {
  // ...
})
```

#### Run a subset of tests

To run a subset of e2e tests:

1. Edit `cypress.json` in `test/e2e/`.

2. Add the `testFiles` option and point it to the specs you want, for example:

   ```json
   "testFiles": ["profiles/network/*"]
   ```

3. Start the tests again using the run script or Makefile.

## Build Docker image

To build a development Docker image:

```make
make docker
```

## Preview API documentation

To work on and preview the generated API documentation:

1. Update the SDK including the OpenAPI specification:

   ```make
   make sdk
   ```

2. Run the preview server for API documentation:

   ```make
   make docs/api
   ```

3. Run the preview server for Swagger documentation:

   ```make
   make docs/swagger
   ```
