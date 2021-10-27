---
id: configuring
title: Configuring
---

This document describes _how_ the service can be configured. For a documentation
on all configuration values head over to the
[configuration reference](reference/configuration.md).

## Configuration File

Files in the format of JSON, YAML and TOML are supported. When starting Kratos
you specify the path to the config file via the `--config` flag. For more
in-depth info run `kratos help`. We recommend you name your configuration file
`kratos.(json|yml|yaml|toml)`. This will enable automatic validation and code
completion in VSCode, JetBrains' IDEs and other tools that utilise
[schemastore.org](https://www.schemastore.org/json/) JSON schemas. Alternatively
you can add the schema mapping manually as well. It is recommended to prevent
leaking of secrets by setting sensitive values via environmental variables as
described in the following section.

## Environmental Variables

Environmental variables take precedence over config file values. Nested paths
get mapped to config values by putting an underscore `_` between every level,
e.g. `selfservice.flows.settings.ui_url` becomes
`SELFSERVICE_FLOWS_SETTINGS_UI_URL=<value>`.

Please note that there are some caveats when using env vars
[documented here](https://www.ory.sh/docs/ecosystem/configuring).
