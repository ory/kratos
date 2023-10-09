#!/bin/sh

# Renders the config schema to a TypeScript file that can be used in Cypress tests.

set -euxo pipefail

dir=$(realpath $(dirname "${BASH_SOURCE[0]}"))

ory_x_version="$(cd $dir/../..; go list -f '{{.Version}}' -m github.com/ory/x)"

curl -s https://raw.githubusercontent.com/ory/x/$ory_x_version/otelx/config.schema.json > .tracing-config.schema.json

sed "s!ory://tracing-config!.tracing-config.schema.json!g;" ../../embedx/config.schema.json | npx json2ts --strictIndexSignatures > cypress/support/config.d.ts

rm .tracing-config.schema.json
