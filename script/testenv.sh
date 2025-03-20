#!/usr/bin/env bash

docker rm -f kratos_test_database_mysql kratos_test_database_postgres kratos_test_database_cockroach kratos_test_hydra || true
docker run --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:8.0
docker run --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:14 postgres -c log_statement=all
docker run --name kratos_test_database_cockroach -p 3446:26257 -p 3447:8080 -d cockroachdb/cockroach:v22.2.6 start-single-node --insecure
docker run --name kratos_test_hydra -p 4444:4444 -p 4445:4445 -d -e DSN=memory -e URLS_SELF_ISSUER=http://localhost:4444/ -e URLS_LOGIN=http://localhost:4446/login -e URLS_CONSENT=http://localhost:4446/consent oryd/hydra:v2.0.2 serve all --dev
docker pull oryd/hydra:v2.2.0@sha256:6c0f9195fe04ae16b095417b323881f8c9008837361160502e11587663b37c09

source script/test-envs.sh
