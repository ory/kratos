#!/bin/bash

set -euxo pipefail

ory_x_version="$(go list -f '{{.Version}}' -m github.com/ory/x)"

sed "s!ory://tracing-config!https://raw.githubusercontent.com/ory/x/$ory_x_version/tracing/config.schema.json!g;
s!ory://logging-config!https://raw.githubusercontent.com/ory/x/$ory_x_version/logrusx/config.schema.json!g" embedx/config.schema.json > .schema/config.schema.json

git add .schema/config.schema.json

if ! git diff --exit-code .schema/config.schema.json
then
  git commit -m "autogen: render config schema"
  git push
fi
