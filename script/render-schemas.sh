#!/bin/sh

set -euxo pipefail

ory_x_version="$(go list -f '{{.Version}}' -m github.com/ory/x)"

sed "s!ory://tracing-config!https://raw.githubusercontent.com/ory/x/$ory_x_version/otelx/config.schema.json!g;" embedx/config.schema.json > .schemastore/config.schema.json

git config user.email "60093411+ory-bot@users.noreply.github.com"
git config user.name "ory-bot"

git add embedx/config.schema.json
git commit -m "autogen: render config schema" || true
