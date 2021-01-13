---
id: kratos-hashers-argon2-load-test
title: kratos hashers argon2 load-test
description: kratos hashers argon2 load-test
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos hashers argon2 load-test

### Synopsis

```
kratos hashers argon2 load-test <authentication-requests-per-minute> [flags]
```

### Options

```
  -c, --config strings                Path to one or more .json, .yaml, .yml, .toml config files. Values are loaded in the order provided, meaning that the last config file overwrites values from the previous config file.
      --dedicated-memory byte_size    Amount of memory dedicated for password hashing. Kratos will try to not consume more memory. (default 4.00GB)
      --expected-deviation duration   Expected deviation of the time a hashing operation (~login request) takes. (default 500ms)
  -f, --format string                 Set the output format. One of table, json, and json-pretty. (default "default")
  -h, --help                          help for load-test
      --iterations uint32             Number of iterations to start probing at. (default 1)
      --key-length uint32             Length of the key in bytes. (default 32)
      --memory byte_size              Memory to use. (default 512.00MB)
      --minimal-duration duration     Minimal duration a hashing operation (~login request) takes. (default 500ms)
      --parallelism uint8             Number of threads to use. (default 72)
  -q, --quiet                         Be quiet with output printing.
      --salt-length uint32            Length of the salt in bytes. (default 16)
```

### SEE ALSO

- [kratos hashers argon2](kratos-hashers-argon2) -
