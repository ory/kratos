#!/usr/bin/env bash

echo "Running Ory Kratos E2E Tests..."
echo ""
set -euxo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

make .bin/hydra

export PATH=.bin:$PATH
export KRATOS_PUBLIC_URL=http://localhost:4433/
export KRATOS_BROWSER_URL=http://localhost:4433/
export KRATOS_ADMIN_URL=http://localhost:4434/
export KRATOS_UI_URL=http://localhost:4456/
export KRATOS_UI_REACT_URL=http://localhost:4458/
export KRATOS_UI_REACT_NATIVE_URL=http://localhost:19006/
export LOG_LEAK_SENSITIVE_VALUES=true
export DEV_DISABLE_API_FLOW_ENFORCEMENT=true
export COOKIE_SECRET=kweifawskf23weas
export CSRF_COOKIE_NAME=node_csrf_token
export CSRF_COOKIE_SECRET=lkaw9oe8isedrhq2

base=$(pwd)

setup=yes
dev=no
nokill=no
cleanup=no
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
  --cleanup)
    cleanup=yes
    shift # past argument=value
    ;;
  esac
done

cleanup() {
    killall node || true
    killall modd || true
    killall webhook || true
    killall hydra || true
    killall hydra-login-consent || true
    killall hydra-kratos-login-consent || true
    docker kill kratos_test_hydra || true
}

prepare() {
  echo "::group::prepare"
  if [[ "${nokill}" == "no" ]]; then
    cleanup
  fi

  if [ -z ${TEST_DATABASE_POSTGRESQL+x} ]; then
    docker rm -f kratos_test_database_mysql kratos_test_database_postgres kratos_test_database_cockroach || true
    docker run --name kratos_test_database_mysql -p 3444:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql:8.0
    docker run --name kratos_test_database_postgres -p 3445:5432 -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=postgres -d postgres:14 postgres -c log_statement=all
    docker run --name kratos_test_database_cockroach -p 3446:26257 -d cockroachdb/cockroach:v22.2.6 start-single-node --insecure

    export TEST_DATABASE_MYSQL="mysql://root:secret@(localhost:3444)/mysql?parseTime=true&multiStatements=true"
    export TEST_DATABASE_POSTGRESQL="postgres://postgres:secret@localhost:3445/postgres?sslmode=disable"
    export TEST_DATABASE_COCKROACHDB="cockroach://root@localhost:3446/defaultdb?sslmode=disable"
  fi

  if [ -z ${NODE_UI_PATH+x} ]; then
    node_ui_dir="$(mktemp -d -t ci-XXXXXXXXXX)/kratos-selfservice-ui-node"
    git clone --depth 1 --branch master https://github.com/ory/kratos-selfservice-ui-node.git "$node_ui_dir"
    (cd "$node_ui_dir" && npm i --legacy-peer-deps && npm run build)
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
  nc -zv localhost 4444 && exit 1
  nc -zv localhost 4445 && exit 1
  nc -zv localhost 4446 && exit 1
  nc -zv localhost 4455 && exit 1
  nc -zv localhost 19006 && exit 1
  nc -zv localhost 4456 && exit 1
  nc -zv localhost 4458 && exit 1
  nc -zv localhost 4744 && exit 1
  nc -zv localhost 4745 && exit 1

  (
    cd "$rn_ui_dir"
    KRATOS_URL=http://localhost:4433 CI=1 npm run web \
      >"${base}/test/e2e/rn-profile-app.e2e.log" 2>&1 &
  )

  hydra serve all -c test/e2e/hydra.yml --dev >"${base}/test/e2e/hydra.e2e.log" 2>&1 &

  (cd test/e2e; npm run wait-on -- -l -t 300000 http-get://localhost:4445/health/alive)

  hydra_client=$(hydra create oauth2-client \
    --endpoint http://localhost:4445 \
    --grant-type authorization_code --grant-type refresh_token \
    --response-type code --response-type id_token \
    --scope openid --scope offline \
    --redirect-uri http://localhost:4455/self-service/methods/oidc/callback/hydra \
    --format json)
  export OIDC_HYDRA_CLIENT_ID=$(jq -r '.client_id' <<< "$hydra_client" )
  export OIDC_HYDRA_CLIENT_SECRET=$(jq -r '.client_secret' <<< "$hydra_client" )

  google_client=$(hydra create oauth2-client \
    --endpoint http://localhost:4445 \
    --grant-type authorization_code --grant-type refresh_token \
    --response-type code --response-type id_token \
    --scope openid --scope offline \
    --redirect-uri http://localhost:4455/self-service/methods/oidc/callback/google \
    --format json)
  export OIDC_GOOGLE_CLIENT_ID=$(jq -r '.client_id' <<< "$google_client" )
  export OIDC_GOOGLE_CLIENT_SECRET=$(jq -r '.client_secret' <<< "$google_client" )

  github_client=$(hydra create oauth2-client \
    --endpoint http://localhost:4445 \
    --grant-type authorization_code --grant-type refresh_token \
    --response-type code --response-type id_token \
    --scope openid --scope offline \
    --redirect-uri http://localhost:4455/self-service/methods/oidc/callback/github \
    --format json)
  export OIDC_GITHUB_CLIENT_ID=$(jq -r '.client_id' <<< "$github_client" )
  export OIDC_GITHUB_CLIENT_SECRET=$(jq -r '.client_secret' <<< "$github_client" )

  (
    cd test/e2e/hydra-login-consent
    go build .
    PORT=4446 HYDRA_ADMIN_URL=http://localhost:4445 ./hydra-login-consent >"${base}/test/e2e/hydra-ui.e2e.log" 2>&1 &
  )

  # Spin up another Hydra instance with the express node app used as the login UI for kratos-hydra OIDC provider tests
  DSN=memory SERVE_PUBLIC_PORT=4744 \
    SERVE_ADMIN_PORT=4745 \
    URLS_SELF_ISSUER=http://localhost:4744 \
    LOG_LEVEL=trace \
    URLS_LOGIN=http://localhost:4455/login \
    URLS_CONSENT=http://localhost:4746/consent \
    SECRETS_SYSTEM="[\"1234567890123456789012345678901\"]" \
    hydra serve all --dev >"${base}/test/e2e/hydra-kratos.e2e.log" 2>&1 &

  (cd test/e2e; npm run wait-on -- -l -t 300000 http-get://127.0.0.1:4745/health/alive)

  dummy_client=$(hydra create oauth2-client \
    --endpoint http://localhost:4745 \
    --token-endpoint-auth-method client_secret_basic \
    --grant-type authorization_code --grant-type refresh_token \
    --response-type code --response-type id_token \
    --scope openid --scope offline --scope email --scope website \
    --redirect-uri http://localhost:5555/callback \
    --redirect-uri https://ory-network-httpbin-ijakee5waq-ez.a.run.app/anything \
    --format json)
  export CYPRESS_OIDC_DUMMY_CLIENT_ID=$(jq -r '.client_id' <<< "$dummy_client" )
  export CYPRESS_OIDC_DUMMY_CLIENT_SECRET=$(jq -r '.client_secret' <<< "$dummy_client" )

  (
    cd test/e2e/hydra-kratos-login-consent
    go build .
    PORT=4746 HYDRA_ADMIN_URL=http://localhost:4745 ./hydra-kratos-login-consent >"${base}/test/e2e/hydra-kratos-ui.e2e.log" 2>&1 &
  )

  (
    cd "$node_ui_dir"
    PORT=4456 SECURITY_MODE=cookie npm run start \
      >"${base}/test/e2e/ui-node.e2e.log" 2>&1 &
  )

  if [ -z ${REACT_UI_PATH+x} ]; then
    (
      cd "$react_ui_dir"
      NEXT_PUBLIC_KRATOS_PUBLIC_URL=http://localhost:4433 npm run build
      NEXT_PUBLIC_KRATOS_PUBLIC_URL=http://localhost:4433 npm run start -- --hostname 127.0.0.1 --port 4458 \
        >"${base}/test/e2e/react-iu.e2e.log" 2>&1 &
    )
  else
    (
      cd "$react_ui_dir"
      PORT=4458 NEXT_PUBLIC_KRATOS_PUBLIC_URL=http://localhost:4433 npm run dev \
        >"${base}/test/e2e/react-iu.e2e.log" 2>&1 &
    )
  fi

  (
    cd test/e2e/proxy
    PORT=4455 npm run start \
      >"${base}/test/e2e/proxy.e2e.log" 2>&1 &
  )

  # Make the environment available to Playwright
  env | grep KRATOS_                         >  test/e2e/playwright/playwright.env
  env | grep TEST_DATABASE_                  >> test/e2e/playwright/playwright.env
  env | grep OIDC_                           >> test/e2e/playwright/playwright.env
  env | grep CYPRESS_                        >> test/e2e/playwright/playwright.env
  echo LOG_LEAK_SENSITIVE_VALUES=true        >> test/e2e/playwright/playwright.env
  echo DEV_DISABLE_API_FLOW_ENFORCEMENT=true >> test/e2e/playwright/playwright.env

  echo "::endgroup::"
}

run() {
  echo "::group::run-prep"
  killall modd || true
  killall kratos || true

  export DSN=${1}

  nc -zv localhost 4434 && (echo "Port 4434 unavailable, used by" ; lsof -i:4434 ; exit 1)
  nc -zv localhost 4433 && (echo "Port 4433 unavailable, used by" ; lsof -i:4433 ; exit 1)

  ls -la .
  for profile in code email mobile oidc recovery recovery-mfa verification mfa mfa-optional spa network passwordless passkey webhooks oidc-provider oidc-provider-mfa two-steps; do
    go tool yq ea '. as $item ireduce ({}; . * $item )' test/e2e/profiles/kratos.base.yml "test/e2e/profiles/${profile}/.kratos.yml" > test/e2e/kratos.${profile}.yml
    cat "test/e2e/kratos.${profile}.yml" | envsubst | sponge "test/e2e/kratos.${profile}.yml"
  done
  cp test/e2e/kratos.email.yml test/e2e/kratos.generated.yml

  (go tool modd -f test/e2e/modd.conf >"${base}/test/e2e/kratos.e2e.log" 2>&1 &)

  npm run wait-on -- -l -t 300000 http-get://127.0.0.1:4434/health/ready \
    http-get://127.0.0.1:4444/.well-known/openid-configuration \
    http-get://127.0.0.1:4455/health/ready \
    http-get://127.0.0.1:4445/health/ready \
    http-get://127.0.0.1:4446/ \
    http-get://127.0.0.1:4456/health/alive \
    http-get://127.0.0.1:19006/ \
    http-get://127.0.0.1:4437/mail \
    http-get://127.0.0.1:4458/ \
    http-get://127.0.0.1:4459/health
  echo "::endgroup::"

  if [[ $dev == "yes" ]]; then
    (cd test/e2e; npm run test:watch --)
  else
    if [ -z ${CYPRESS_RECORD_KEY+x} ]; then
      (cd test/e2e; npm run test --)
    else
      (cd test/e2e; npm run test -- --record --tag "${2}" )
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

if [[ "${cleanup}" == "yes" ]]; then
  cleanup
  exit 0
fi

export TEST_DATABASE_SQLITE="sqlite:///$(mktemp -d -t ci-XXXXXXXXXX)/db.sqlite?_fk=true"
export TEST_DATABASE_MEMORY="memory"

case "${1:-default}" in
sqlite)
  echo "Database set up at: $TEST_DATABASE_SQLITE"
  dsn="${TEST_DATABASE_SQLITE}"
  db="sqlite"
  ;;

mysql)
  dsn="${TEST_DATABASE_MYSQL}"
  db="mysql"
  ;;

postgres)
  dsn="${TEST_DATABASE_POSTGRESQL}"
  db="postgres"
  ;;

cockroach)
  dsn="${TEST_DATABASE_COCKROACHDB}"
  db="cockroach"
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

run "${dsn}" "${db}"
