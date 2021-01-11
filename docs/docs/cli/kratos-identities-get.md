---
id: kratos-identities-get
title: kratos identities get
description: kratos identities get Get one or more identities by ID
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->
## kratos identities get

Get one or more identities by ID

### Synopsis

This command gets all the details about an identity. To get an identity by some selector, e.g. the recovery email address, use the list command in combination with jq.

We have to admit, this is not easy if you don't speak jq fluently. What about opening an issue and telling us what predefined selectors you want to have? https://github.com/ory/kratos/issues/new/choose


```
kratos identities get <id-0 [id-1 ...]> [flags]
```

### Examples

```
To get the identities with the recovery email address at the domain "ory.sh", run:

	$ kratos identities get $(kratos identities list --format json | jq -r 'map(select(.recovery_addresses[].value | endswith("@ory.sh"))) | .[].id')
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -e, --endpoint string   The URL of ORY Kratos' Admin API. Alternatively set using the KRATOS_ADMIN_URL environmental variable.
  -f, --format string     Set the output format. One of table, json, and json-pretty.
  -q, --quiet             Prints only IDs, one per line. Takes precedence over --format.
```

### SEE ALSO

* [kratos identities](kratos-identities)	 - Tools to interact with remote identities

