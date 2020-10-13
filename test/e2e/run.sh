#!/bin/bash

set -euxo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/../.."

make .bin/hydra
make .bin/yq

export PATH=.bin:$PATH
export KRATOS_PUBLIC_URL=http://127.0.0.1:4433/
export KRATOS_BROWSER_URL=http://127.0.0.1:4433/
export KRATOS_ADMIN_URL=http://127.0.0.1:4434/
export KRATOS_UI_URL=http://127.0.0.1:4456/
export LOG_LEAK_SENSITIVE_VALUES=true

if [ -z ${TEST_DATABASE_POSTGRESQL+x} ]; then
  docker rm -f kratos_test_database_mysql kratos_test_database_postgres kratos_test_database_cockroach || true
  docker run --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:5.7
  docker run --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:9.6 postgres -c log_statement=all
  docker run --name kratos_test_database_cockroach -p 3446:26257 -d cockroachdb/cockroach:v20.1.1 start --insecure

  export TEST_DATABASE_MYSQL="mysql://root:secret@(127.0.0.1:3444)/mysql?parseTime=true&multiStatements=true"
  export TEST_DATABASE_POSTGRESQL="postgres://postgres:secret@127.0.0.1:3445/postgres?sslmode=disable"
  export TEST_DATABASE_COCKROACHDB="cockroach://root@127.0.0.1:3446/defaultdb?sslmode=disable"
fi

! nc -zv 127.0.0.1 4434
! nc -zv 127.0.0.1 4433
! nc -zv 127.0.0.1 4446
! nc -zv 127.0.0.1 4456
! nc -zv 127.0.0.1 4455

base=$(pwd)

if [ -z ${KRATOS_APP_PATH+x} ]; then
  dir="$(mktemp -d -t ci-XXXXXXXXXX)/kratos-selfservice-ui-node"
  git clone git@github.com:ory/kratos-selfservice-ui-node.git "$dir"
  (cd "$dir"; npm i && npm run build)
else
  dir="${KRATOS_APP_PATH}"
fi

(cd test/e2e/proxy; npm i)

kratos=./test/e2e/.bin/kratos
go build -tags sqlite -o $kratos .

if [ -z ${CI+x} ]; then
  docker rm mailslurper hydra hydra-ui -f || true
  docker run --name mailslurper -p 4436:4436 -p 4437:4437 -p 1025:1025 oryd/mailslurper:latest-smtps > "${base}/test/e2e/mailslurper.e2e.log" 2>&1 &
fi

dev=no
for i in "$@"
do
case $i in
    --dev)
    dev=yes
    shift # past argument=value
    ;;
esac
done

run() {
  profile=$2
  killall kratos || true
  killall node || true
  killall hydra || true
  killall hydra-login-consent || true

  DSN=memory URLS_SELF_ISSUER=http://127.0.0.1:4444 \
    LOG_LEVEL=trace \
    URLS_LOGIN=http://127.0.0.1:4446/login \
    URLS_CONSENT=http://127.0.0.1:4446/consent \
    hydra serve all --dangerous-force-http > "${base}/test/e2e/hydra.e2e.log" 2>&1 &

  hydra clients create \
    --endpoint http://127.0.0.1:4445 \
    --id kratos-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://127.0.0.1:4455/self-service/methods/oidc/callback/hydra

  hydra clients create \
    --endpoint http://127.0.0.1:4445 \
    --id google-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://127.0.0.1:4455/self-service/methods/oidc/callback/google

  hydra clients create \
    --endpoint http://127.0.0.1:4445 \
    --id github-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://127.0.0.1:4455/self-service/methods/oidc/callback/github

  if [ -z ${KRATOS_APP_PATH+x} ]; then
    (cd "$dir"; PORT=4456 SECURITY_MODE=cookie npm run serve \
      > "${base}/test/e2e/secureapp.e2e.log" 2>&1 &)
  else
    (cd "$dir"; PORT=4456 SECURITY_MODE=cookie npm run start \
     > "${base}/test/e2e/secureapp.e2e.log" 2>&1 &)
  fi

  (cd test/e2e/proxy; PORT=4455 npm run start \
   > "${base}/test/e2e/proxy.e2e.log" 2>&1 &)

  (cd test/e2e/hydra-login-consent; \
    go build . && \
    PORT=4446 HYDRA_ADMIN_URL=http://127.0.0.1:4445 ./hydra-login-consent > "${base}/test/e2e/hydra-ui.e2e.log" 2>&1 &)

  export DSN=${1}
  if [ "$DSN" != "memory" ]; then
    $kratos migrate sql -e --yes
  fi

  yq merge test/e2e/profiles/kratos.base.yml "test/e2e/profiles/${profile}/.kratos.yml" > test/e2e/kratos.generated.yml
  ($kratos serve --dev -c test/e2e/kratos.generated.yml > "${base}/test/e2e/kratos.${profile}.e2e.log" 2>&1 &)

  npm run wait-on -- -t 10000 http-get://127.0.0.1:4434/health/ready \
    http-get://127.0.0.1:4455/health \
    http-get://127.0.0.1:4445/health/ready \
    http-get://127.0.0.1:4446/ \
    http-get://127.0.0.1:4455/ \
    http-get://127.0.0.1:4456/ \
    http-get://127.0.0.1:4437/mail

  if [[ $dev = "yes" ]]; then
    npm run test:watch -- --config integrationFolder="test/e2e/cypress/integration/profiles/$profile"
  else
    npm run test -- --config integrationFolder="test/e2e/cypress/integration/profiles/$profile"
  fi
}

usage() {
    echo $"This script runs the e2e tests.

To run the tests just pick a database name:

  $0 <database>

  Supported databases are 'sqlite', 'mysql', 'postgres', 'cockroach':

    $0 sqlite
    $0 mysql
    $0 postgres
    $0 cockroach
    ...

  If you are using a database other than SQLite, you need to set
  an environment variable that points to it:

    export TEST_DATABASE_MYSQL=...
    export TEST_DATABASE_POSTGRESQL=...
    export TEST_DATABASE_COCKROACHDB=...
    $0 <database>

  The Makefile has a helper for that which uses Docker to start the
  databases:

    make test-resetdb
    source scripts/test-envs.sh
    $0 <database>

To run e2e tests in dev mode (useful for writing them), run:

  $0 --dev <database> <profile>

  Supported profiles are 'email', 'verification', 'oidc', 'recovery':

    $0 --dev <database> email
    $0 --dev <database> verification
    ...

If you are making changes to the kratos-selfservice-ui-node
project as well, point the 'KRATOS_APP_PATH' environment variable to
the path where the kratos-selfservice-ui-node project is checked out:

  export KRATOS_APP_PATH=$HOME/workspace/kratos-selfservice-ui-node
  $0 ..."
}

if [[ $dev = "yes" ]]; then
  if [ -z ${2+x} ]; then
    usage
    exit 1
  fi
fi

export TEST_DATABASE_SQLITE="sqlite:///$(mktemp -d -t ci-XXXXXXXXXX)/db.sqlite?_fk=true"
export TEST_DATABASE_MEMORY="memory"

case "$1" in
        sqlite)
          db="${TEST_DATABASE_SQLITE}"
          ;;

        mysql)
          db="${TEST_DATABASE_MYSQL}"
          ;;

        postgres)
          db="${TEST_DATABASE_POSTGRESQL}"
          ;;

        cockroach)
          db="${TEST_DATABASE_COCKROACHDB}"
          ;;

        *)
            usage
            exit 1
esac

if [[ $dev = "yes" ]]; then
  run "${db}" "$2"
else
  run "${db}" email
  run "${db}" verification
  run "${db}" oidc
  run "${db}" recovery
fi
