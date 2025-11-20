#!/bin/sh

# Renders the config schema to a TypeScript file that can be used in Cypress tests.

set -euxo pipefail

dir=$(realpath $(dirname "${BASH_SOURCE[0]}"))

ory_x_version="$(cd $dir/../..; go list -f '{{.Version}}' -m github.com/ory/x)"

curl --retry 7 --retry-connrefused -s https://raw.githubusercontent.com/ory/x/$ory_x_version/otelx/config.schema.json > $dir/.tracing-config.schema.json

(cd $dir; sed "s!ory://tracing-config!.tracing-config.schema.json!g;" $dir/../../embedx/config.schema.json | npx json2ts --strictIndexSignatures > $dir/shared/config.d.ts)

rm $dir/.tracing-config.schema.json

(cd $dir/../..; make format)
