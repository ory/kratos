#!/bin/bash

set -euxo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

make .bin/hydra
make .bin/yq
make .bin/modd

export PATH=.bin:$PATH
export KRATOS_PUBLIC_URL=http://localhost:4433/
export KRATOS_BROWSER_URL=http://localhost:4433/
export KRATOS_ADMIN_URL=http://localhost:4434/
export KRATOS_UI_URL=http://127.0.0.1:4456/
export KRATOS_UI_REACT_URL=http://127.0.0.1:4458/
export KRATOS_UI_REACT_NATIVE_URL=http://127.0.0.1:4457/
export LOG_LEAK_SENSITIVE_VALUES=true
export DEV_DISABLE_API_FLOW_ENFORCEMENT=true

base=$(pwd)

setup=yes
dev=no
nokill=no
for i in "$@"; do
  case $i in
  --no-kill)
    nokill=yes
    shift # past argument=value
    ;;
  --only-setup)
    setup=only
    shift # past argument=value
    ;;
  --no-setup)
    setup=no
    shift # past argument=value
    ;;
  --dev)
    dev=yes
    shift # past argument=value
    ;;
  esac
done

prepare() {
  if [[ "${nokill}" == "no" ]]; then
    killall node || true
    killall modd || true
    killall hydra || true
    killall hydra-login-consent || true
  fi

  if [ -z ${TEST_DATABASE_POSTGRESQL+x} ]; then
    docker rm -f kratos_test_database_mysql kratos_test_database_postgres kratos_test_database_cockroach || true
    docker run --platform linux/amd64 --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:5.7
    docker run --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:9.6 postgres -c log_statement=all
    docker run --name kratos_test_database_cockroach -p 3446:26257 -d cockroachdb/cockroach:v20.2.4 start-single-node --insecure

    export TEST_DATABASE_MYSQL="mysql://root:secret@(localhost:3444)/mysql?parseTime=true&multiStatements=true"
    export TEST_DATABASE_POSTGRESQL="postgres://postgres:secret@localhost:3445/postgres?sslmode=disable"
    export TEST_DATABASE_COCKROACHDB="cockroach://root@localhost:3446/defaultdb?sslmode=disable"
  fi

  if [ -z ${NODE_UI_PATH+x} ]; then
    node_ui_dir="$(mktemp -d -t ci-XXXXXXXXXX)/kratos-selfservice-ui-node"
    git clone --depth 1 --branch master https://github.com/ory/kratos-selfservice-ui-node.git "$node_ui_dir"
    (cd "$node_ui_dir" && npm i && npm run build)
  else
    node_ui_dir="${NODE_UI_PATH}"
  fi

  if [ -z ${RN_UI_PATH+x} ]; then
    rn_ui_dir="$(mktemp -d -t ci-XXXXXXXXXX)/kratos-selfservice-ui-react-native"
    git clone --depth 1 --branch master https://github.com/ory/kratos-selfservice-ui-react-native.git "$rn_ui_dir"
    (cd "$rn_ui_dir" && npm i)
  else
    rn_ui_dir="${RN_UI_PATH}"
  fi

  if [ -z ${REACT_UI_PATH+x} ]; then
    react_ui_dir="$(mktemp -d -t ci-XXXXXXXXXX)/ory/kratos-selfservice-ui-react-nextjs"
    git clone --depth 1 --branch master https://github.com/ory/kratos-selfservice-ui-react-nextjs.git "$react_ui_dir"
    (cd "$react_ui_dir" && npm i)
  else
    react_ui_dir="${REACT_UI_PATH}"
  fi

  (
    rm test/e2e/proxy.json || true
    echo '"express"' > test/e2e/proxy.json
    cd test/e2e/proxy
    npm i
  )

  if [ -z ${CI+x} ]; then
    docker rm mailslurper hydra hydra-ui -f || true
    docker run --name mailslurper -p 4436:4436 -p 4437:4437 -p 1025:1025 oryd/mailslurper:latest-smtps > "${base}/test/e2e/mailslurper.e2e.log" 2>&1 &
  fi

  # Check if any ports that we need are open already
  ! nc -zv localhost 4446
  ! nc -zv localhost 4455
  ! nc -zv localhost 4456
  ! nc -zv localhost 4457
  ! nc -zv localhost 4458

  (
    cd "$rn_ui_dir"
    npm i -g expo-cli
    WEB_PORT=4457 KRATOS_URL=http://localhost:4433 npm run web -- --non-interactive \
      >"${base}/test/e2e/rn-profile-app.e2e.log" 2>&1 &
  )

  DSN=memory URLS_SELF_ISSUER=http://localhost:4444 \
    LOG_LEVEL=trace \
    URLS_LOGIN=http://localhost:4446/login \
    URLS_CONSENT=http://localhost:4446/consent \
    hydra serve all --dangerous-force-http >"${base}/test/e2e/hydra.e2e.log" 2>&1 &

  (cd test/e2e; npm run wait-on -- -l -t 300000 http-get://127.0.0.1:4445/health/alive)

  hydra clients delete \
    --endpoint http://localhost:4445 \
    kratos-client google-client github-client || true

  hydra clients create \
    --endpoint http://localhost:4445 \
    --id kratos-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://localhost:4455/self-service/methods/oidc/callback/hydra

  hydra clients create \
    --endpoint http://localhost:4445 \
    --id google-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://localhost:4455/self-service/methods/oidc/callback/google

  hydra clients create \
    --endpoint http://localhost:4445 \
    --id github-client \
    --secret kratos-secret \
    --grant-types authorization_code,refresh_token \
    --response-types code,id_token \
    --scope openid,offline \
    --callbacks http://localhost:4455/self-service/methods/oidc/callback/github

  if [ -z ${NODE_UI_PATH+x} ]; then
    (
      cd "$node_ui_dir"
      PORT=4456 SECURITY_MODE=cookie npm run serve \
        >"${base}/test/e2e/ui-node.e2e.log" 2>&1 &
    )
  else
    (
      cd "$node_ui_dir"
      PORT=4456 SECURITY_MODE=cookie npm run start \
        >"${base}/test/e2e/ui-node.e2e.log" 2>&1 &
    )
  fi

  if [ -z ${REACT_UI_PATH+x} ]; then
    (
      cd "$react_ui_dir"
      ORY_KRATOS_URL=http://localhost:4433 npm run build
      ORY_KRATOS_URL=http://localhost:4433 npm run start -- --hostname 0.0.0.0 --port 4458 \
        >"${base}/test/e2e/react-iu.e2e.log" 2>&1 &
    )
  else
    (
      cd "$react_ui_dir"
      PORT=4458 ORY_KRATOS_URL=http://localhost:4433 npm run dev \
        >"${base}/test/e2e/react-iu.e2e.log" 2>&1 &
    )
  fi

  (
    cd test/e2e/proxy
    PORT=4455 npm run start \
      >"${base}/test/e2e/proxy.e2e.log" 2>&1 &
  )

  (
    cd test/e2e/hydra-login-consent
    go build . &&
      PORT=4446 HYDRA_ADMIN_URL=http://localhost:4445 ./hydra-login-consent >"${base}/test/e2e/hydra-ui.e2e.log" 2>&1 &
  )
}

run() {
  killall modd || true
  killall kratos || true

  export DSN=${1}

  ! nc -zv localhost 4434
  ! nc -zv localhost 4433

  ls -la .
  for profile in email mobile oidc recovery verification mfa spa; do
    yq merge test/e2e/profiles/kratos.base.yml "test/e2e/profiles/${profile}/.kratos.yml" > test/e2e/kratos.${profile}.yml
    cp test/e2e/kratos.email.yml test/e2e/kratos.generated.yml
  done

  (modd -f test/e2e/modd.conf >"${base}/test/e2e/kratos.e2e.log" 2>&1 &)

  npm run wait-on -- -v -l -t 300000 http-get://127.0.0.1:4434/health/ready \
    http-get://127.0.0.1:4455/health/ready \
    http-get://127.0.0.1:4445/health/ready \
    http-get://127.0.0.1:4446/ \
    http-get://127.0.0.1:4456/health/alive \
    http-get://127.0.0.1:4457/ \
    http-get://127.0.0.1:4437/mail \
    http-get://127.0.0.1:4458/

  if [[ $dev == "yes" ]]; then
    (cd test/e2e; npm run test:watch -- --config integrationFolder="cypress/integration")
  else
    if [ -z ${CYPRESS_RECORD_KEY+x} ]; then
      (cd test/e2e; npm run test -- --config integrationFolder="cypress/integration")
    else
      (cd test/e2e; npm run test -- --record --config integrationFolder="cypress/integration")
    fi
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
    source script/test-envs.sh
    $0 <database>

To run e2e tests in dev mode (useful for writing them), run:

  $0 --dev <database>

To set up all the services without running the tests, use:

  $0 --only-setup

To then run the tests without the set up steps, use:

  $0 --no-setup <database>

To prevent processes from being killed during set up phase, use:
  $0 --no-kill

If you are making changes to the kratos-selfservice-ui-node
project as well, point the 'NODE_UI_PATH' environment variable to
the path where the kratos-selfservice-ui-node project is checked out:

  export NODE_UI_PATH=$HOME/workspace/kratos-selfservice-ui-node
  export RN_UI_PATH=$HOME/workspace/kratos-selfservice-ui-react-native
  export REACT_UI_PATH=$HOME/workspace/kratos-selfservice-ui-react-nextjs
  $0 ..."
}

export TEST_DATABASE_SQLITE="sqlite:///$(mktemp -d -t ci-XXXXXXXXXX)/db.sqlite?_fk=true"
export TEST_DATABASE_MEMORY="memory"

case "${1:-default}" in
sqlite)
  echo "Database set up at: $TEST_DATABASE_SQLITE"
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
  if [[ "${setup}" == "only" ]]; then
    prepare
    exit 0
  else
    exit 1
  fi
  ;;
esac

if [[ "${setup}" == "yes" ]]; then
  prepare
fi
run "${db}"
