---
id: kratos-identities-put
title: kratos identities put
description: kratos identities put Put identities from files or STD_IN
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->
## kratos identities put

Put identities from files or STD_IN

### Synopsis

Put (as in http PUT) identities from files or STD_IN. Files are expected to each contain a single identity. The validity of files can be tested beforehand using `... identities validate`.

```
kratos identities put <file.json [file-2.json [file-3.json] ...]> [flags]
```

### Options

```
  -h, --help   help for put
```

### Options inherited from parent commands

```
  -e, --endpoint string   The upstream admin endpoint URL. Alternatively set using the KRATOS_ADMIN_ENDPOINT environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### SEE ALSO

* [kratos identities](kratos-identities)	 - Tools to interact with remote identities

