#!/bin/bash

docker rm -f kratos_test_database_mysql kratos_test_database_postgres kratos_test_database_cockroach kratos_test_hydra || true
docker run --rm --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:5.7
docker run --rm --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:9.6
docker run --rm --name kratos_test_database_cockroach -p 3446:26257 -d cockroachdb/cockroach:v2.1.6 start --insecure
docker run --name kratos_test_hydra -p 4444:4444 -p 4445:4445 -d -e DSN=memory -e URLS_SELF_ISSUER=http://127.0.0.1:4444 -e URLS_LOGIN=http://127.0.0.1:4446/login -e URLS_CONSENT=http://127.0.0.1:4446/consent oryd/hydra:v1.4.10 serve all --dangerous-force-http

source script/test-envs.sh
