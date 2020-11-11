---
id: kratos-identities
title: kratos identities
description: kratos identities Tools to interact with remote identities
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos identities

Tools to interact with remote identities

### Synopsis

Tools to interact with remote identities

### Options

```
  -e, --endpoint string   The URL of ORY Kratos' Admin API. Alternatively set using the KRATOS_ADMIN_URL environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -h, --help              help for identities
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### Options inherited from parent commands

```
  -c, --config string   Path to config file. Supports .json, .yaml, .yml, .toml. Default is "$HOME/.kratos.(yaml|yml|toml|json)"
```

### SEE ALSO

- [kratos](kratos) -
- [kratos identities delete](kratos-identities-delete) - Delete identities by ID
- [kratos identities get](kratos-identities-get) - Get one or more identities by ID
- [kratos identities import](kratos-identities-import) - Import identities from files or STD_IN
- [kratos identities list](kratos-identities-list) - List identities
- [kratos identities patch](kratos-identities-patch) - Patch identities by ID (not yet implemented)
- [kratos identities validate](kratos-identities-validate) - Validate local identity files
