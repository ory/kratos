---
id: kratos-identities-validate
title: kratos identities validate
description: kratos identities validate Validate local identity files
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos identities validate

Validate local identity files

### Synopsis

This command allows validation of identity files.
It validates against the payload of the API and the identity schema as configured in Kratos.
Identities can be supplied via STD_IN or JSON files containing a single or an array of identities.

```
kratos identities validate <file.json [file-2.json [file-3.json] ...]> [flags]
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
  -e, --endpoint string   The URL of ORY Kratos' Admin API. Alternatively set using the KRATOS_ADMIN_URL environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### SEE ALSO

- [kratos identities](kratos-identities) - Tools to interact with remote identities
