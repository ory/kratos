---
id: kratos-identities-delete
title: kratos identities delete
description: kratos identities delete Delete identities by ID
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos identities delete

Delete identities by ID

### Synopsis

This command deletes one or more identities by ID. To delete an identity by some
selector, e.g. the recovery email address, use the list command in combination
with jq. Example: delete the identity with the recovery email address
"foo@bar.com":

kratos identities delete \$(kratos identities list --format json | jq -r
'map(select(.recovery_addresses[].value == "foo@bar.com")) | .[].id')

I have to admit, this is not easy if you don't speak jq fluently. What about
opening an issue and telling us what predefined selectors you want to have?
https://github.com/ory/kratos/issues/new/choose

```
kratos identities delete <id-0 [id-1 ...]> [flags]
```

### Options

```
  -h, --help   help for delete
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
