---
id: kratos-identities-import
title: kratos identities import
description: kratos identities import Import identities from files or STD_IN
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos identities import

Import identities from files or STD_IN

### Synopsis

Import identities from files or STD_IN.

Files can contain only a single or an array of identities. The validity of files
can be tested beforehand using "... identities validate".

WARNING: Importing credentials is not yet supported.

```
kratos identities import <file.json [file-2.json [file-3.json] ...]> [flags]
```

### Examples

```
$ cat > ./file.json <<EOF
{
    "schema_id": "default";,
    "traits": {
        "email": "foo@example.com";
    }
}
EOF

$ kratos identities import file.json
# Alternatively:
$ cat file.json | kratos identities import
```

### Options

```
  -h, --help   help for import
```

### Options inherited from parent commands

```
  -e, --endpoint string   The URL of Ory Kratos&#39; Admin API. Alternatively set using the KRATOS_ADMIN_URL environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty. (default &#34;default&#34;)
  -q, --quiet             Be quiet with output printing.
```

### SEE ALSO

- [kratos identities](kratos-identities) - Tools to interact with remote
  identities
