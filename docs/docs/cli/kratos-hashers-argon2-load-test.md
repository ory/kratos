---
id: kratos-hashers-argon2-load-test
title: kratos hashers argon2 load-test
description:
  kratos hashers argon2 load-test Simulate the password hashing with a number of
  concurrent requests/minute.
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos hashers argon2 load-test

Simulate the password hashing with a number of concurrent requests/minute.

### Synopsis

Simulates a number of concurrent authentication requests per minute. Gives
statistical data about the measured performance and resource consumption. Can be
used to tune and test the hashing parameters for peak demand situations.

```
kratos hashers argon2 load-test &lt;authentication-requests-per-minute&gt; [flags]
```

### Options

```
  -c, --config strings                Path to one or more .json, .yaml, .yml, .toml config files. Values are loaded in the order provided, meaning that the last config file overwrites values from the previous config file.
      --dedicated-memory byte_size    Amount of memory dedicated for password hashing. Kratos will try to not consume more memory. (default 1.00GB)
      --expected-deviation duration   Expected deviation of the time a hashing operation (~login request) takes. (default 500ms)
  -f, --format string                 Set the output format. One of table, json, and json-pretty. (default &#34;default&#34;)
  -h, --help                          help for load-test
      --iterations uint32             Number of iterations to start probing at. (default 1)
      --key-length uint32             Length of the key in bytes. (default 32)
      --memory byte_size              Memory to use. (default 128.00MB)
      --min-duration duration         Minimal duration a hashing operation (~login request) takes. (default 500ms)
      --parallelism uint8             Number of threads to use. (default 96)
  -q, --quiet                         Be quiet with output printing.
      --salt-length uint32            Length of the salt in bytes. (default 16)
```

### SEE ALSO

- [kratos hashers argon2](kratos-hashers-argon2) -
