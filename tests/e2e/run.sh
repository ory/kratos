#!/bin/bash

set -euxo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/../.."

export KRATOS_PUBLIC_URL=http://127.0.0.1:4433/
export KRATOS_ADMIN_URL=http://127.0.0.1:4434/

! nc -zv 127.0.0.1 4434
! nc -zv 127.0.0.1 4433
! nc -zv 127.0.0.1 4455

base=$(pwd)

if [ -z ${KRATOS_APP_PATH+x} ]; then
  dir="$(mktemp -d -t ci-XXXXXXXXXX)/kratos-selfservice-ui-node"
  git clone git@github.com:ory/kratos-selfservice-ui-node.git "$dir"
  (cd "$dir"; \
    npm i; \
    KRATOS_DIR=${base} make build-sdk; \
    npm run build)
else
  dir="${KRATOS_APP_PATH}"
fi

kratos=./tests/e2e/.bin/kratos
go build -tags sqlite -o $kratos .

if [ -z ${CI+x} ]; then
  docker rm mailslurper -f || true
  docker run --name mailslurper -p 4436:4436 -p 4437:4437 -p 1025:1025 oryd/mailslurper:latest-smtps > "${base}/tests/e2e/mailslurper.e2e.log" 2>&1 &
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

  if [ -z ${KRATOS_APP_PATH+x} ]; then
    (cd "$dir"; PORT=4455 SECURITY_MODE=cookie npm run serve \
      > "${base}/tests/e2e/secureapp.e2e.log" 2>&1 &)
  else
    (cd "$dir"; PORT=4455 SECURITY_MODE=cookie npm run start \
     > "${base}/tests/e2e/secureapp.e2e.log" 2>&1 &)
  fi

  export DSN=${1}
  $kratos migrate sql -e --yes

  yq merge tests/e2e/profiles/kratos.base.yml "tests/e2e/profiles/${profile}/.kratos.yml" > tests/e2e/kratos.generated.yml
  ($kratos serve --dev -c tests/e2e/kratos.generated.yml > "${base}/tests/e2e/kratos.${profile}.e2e.log" 2>&1 &)

  npm run wait-on -- -t 10000 http-get://127.0.0.1:4434/health/ready \
    http-get://127.0.0.1:4455/health \
    http-get://127.0.0.1:4437/mail

  if [[ $dev = "yes" ]]; then
    npm run test:watch -- --config integrationFolder="cypress/integration/profiles/$profile"
  else
    npm run test -- --config integrationFolder="cypress/integration/profiles/$profile"
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

To run e2e tests in dev mode (useful for writing them), run:

  $0 --dev <database> <profile>

  Supported profiles are 'email', 'verify':

    $0 --dev <database> email
    $0 --dev <database> verify
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
case "$1" in
        sqlite)
          if [[ $dev = "yes" ]]; then
            run "${TEST_DATABASE_SQLITE}" "$2"
          else
            run "${TEST_DATABASE_SQLITE}" email
            run "${TEST_DATABASE_SQLITE}" profile
          fi
          ;;

        mysql)
          run "${TEST_DATABASE_MYSQL}"
          ;;

        postgres)
          run "${TEST_DATABASE_POSTGRESQL}"
          ;;

        cockroach)
          run "${TEST_DATABASE_COCKROACHDB}"
          ;;

        *)
            usage
            exit 1
esac
