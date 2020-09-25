---
id: kratos-identities-import
title: kratos identities import
description: kratos identities import import identities from files or STD_IN
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos identities import

import identities from files or STD_IN

### Synopsis

Import identities from files or STD_IN. Files are expected to each contain a
single identity. The validity of files can be tested beforehand using
`... identities validate`. Importing credentials is not yet supported.

```
kratos identities import <file.json [file-2.json [file-3.json] ...]> [flags]
```

### Options

```
  -h, --help   help for import
```

### Options inherited from parent commands

```
  -e, --endpoint string   The upstream admin endpoint URL. Alternatively set using the KRATOS_ADMIN_ENDPOINT environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### SEE ALSO

- [kratos identities](kratos-identities) - Tools to interact with remote
  identities
