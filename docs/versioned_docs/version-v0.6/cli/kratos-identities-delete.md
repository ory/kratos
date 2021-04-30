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
with jq.

We have to admit, this is not easy if you don&#39;t speak jq fluently. What
about opening an issue and telling us what predefined selectors you want to
have? https://github.com/ory/kratos/issues/new/choose

```
kratos identities delete &lt;id-0 [id-1 ...]&gt; [flags]
```

### Examples

```
To delete the identity with the recovery email address &#34;foo@bar.com&#34;, run:

	$ kratos identities delete $(kratos identities list --format json | jq -r &#39;map(select(.recovery_addresses[].value == &#34;foo@bar.com&#34;)) | .[].id&#39;)
```

### Options

```
  -h, --help   help for delete
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
