#!/bin/sh

set -euxo pipefail

schema_version="$(git rev-parse --short HEAD)"

sed "s!ory://tracing-config!https://raw.githubusercontent.com/ory/kratos/$schema_version/oryx/otelx/config.schema.json!g;" embedx/config.schema.json > .schemastore/config.schema.json

git config user.email "60093411+ory-bot@users.noreply.github.com"
git config user.name "ory-bot"

git add embedx/config.schema.json
git commit -m "autogen: render config schema" || true
