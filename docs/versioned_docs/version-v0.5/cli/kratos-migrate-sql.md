---
id: kratos-migrate-sql
title: kratos migrate sql
description: kratos migrate sql Create SQL schemas and apply migration plans
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos migrate sql

Create SQL schemas and apply migration plans

### Synopsis

Run this command on a fresh SQL installation and when you upgrade ORY Kratos to a new minor version.

It is recommended to run this command close to the SQL instance (e.g. same subnet) instead of over the public internet.
This decreases risk of failure and decreases time required.

You can read in the database URL using the -e flag, for example:
export DSN=...
kratos migrate sql -e

### WARNING

Before running this command on an existing database, create a back up!

```
kratos migrate sql <database-url> [flags]
```

### Options

```
  -h, --help            help for sql
  -e, --read-from-env   If set, reads the database connection string from the environment variable DSN or config file key dsn.
  -y, --yes             If set all confirmation requests are accepted without user interaction.
```

### Options inherited from parent commands

```
  -c, --config string   Path to config file. Supports .json, .yaml, .yml, .toml. Default is "$HOME/.kratos.(yaml|yml|toml|json)"
```

### SEE ALSO

- [kratos migrate](kratos-migrate) - Various migration helpers
