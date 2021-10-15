---
id: kratos-hashers-argon2-hash
title: kratos hashers argon2 hash
description:
  kratos hashers argon2 hash Hash a list of passwords for benchmarking the
  hashing parameters
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos hashers argon2 hash

Hash a list of passwords for benchmarking the hashing parameters

```
kratos hashers argon2 hash &lt;password1&gt; [&lt;password2&gt; ...] [flags]
```

### Options

```
  -c, --config strings                Path to one or more .json, .yaml, .yml, .toml config files. Values are loaded in the order provided, meaning that the last config file overwrites values from the previous config file.
      --dedicated-memory byte_size    Amount of memory dedicated for password hashing. Kratos will try to not consume more memory. (default 1.00GB)
      --expected-deviation duration   Expected deviation of the time a hashing operation (~login request) takes. (default 500ms)
  -h, --help                          help for hash
      --iterations uint32             Number of iterations to start probing at. (default 1)
      --key-length uint32             Length of the key in bytes. (default 32)
      --memory byte_size              Memory to use. (default 128.00MB)
      --min-duration duration         Minimal duration a hashing operation (~login request) takes. (default 500ms)
      --parallel                      Run all hashing operations in parallel.
      --parallelism uint8             Number of threads to use. (default 96)
      --salt-length uint32            Length of the salt in bytes. (default 16)
```

### SEE ALSO

- [kratos hashers argon2](kratos-hashers-argon2) -
