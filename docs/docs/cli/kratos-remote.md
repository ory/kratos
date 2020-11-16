---
id: kratos-remote
title: kratos remote
description:
  kratos remote Helpers and management for remote ORY Kratos instances
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos remote

Helpers and management for remote ORY Kratos instances

### Synopsis

Helpers and management for remote ORY Kratos instances

### Options

```
  -e, --endpoint string   The URL of ORY Kratos' Admin API. Alternatively set using the KRATOS_ADMIN_URL environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -h, --help              help for remote
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### Options inherited from parent commands

```
  -c, --config string   Path to config file. Supports .json, .yaml, .yml, .toml. Default is "$HOME/.kratos.(yaml|yml|toml|json)"
```

### SEE ALSO

- [kratos](kratos) -
- [kratos remote status](kratos-remote-status) - Print the alive and readiness
  status of a ORY Kratos instance
- [kratos remote version](kratos-remote-version) - Print the version of an ORY
  Kratos instance
