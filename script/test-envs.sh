#! /bin/bash

export TEST_DATABASE_MYSQL="mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true&multiStatements=true"
export TEST_DATABASE_POSTGRESQL="postgres://postgres:secret@127.0.0.1:3445/postgres?sslmode=disable"
export TEST_DATABASE_COCKROACHDB="cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable"
export TEST_SELFSERVICE_OIDC_HYDRA_ADMIN=http://127.0.0.1:4445
export TEST_SELFSERVICE_OIDC_HYDRA_PUBLIC=http://127.0.0.1:4444
export TEST_SELFSERVICE_OIDC_HYDRA_INTEGRATION_ADDR=http://127.0.0.1:4446
