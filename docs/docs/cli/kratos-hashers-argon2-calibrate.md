---
id: kratos-hashers-argon2-calibrate
title: kratos hashers argon2 calibrate
description: kratos hashers argon2 calibrate Calibrate the values for Argon2 so the hashing operation takes the desired time.
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos hashers argon2 calibrate

Calibrate the values for Argon2 so the hashing operation takes the desired time.

### Synopsis

Calibrate the configuration values for Argon2 by probing the execution time.
Note that the values depend on the machine you run the hashing on.
When choosing the desired time, UX is in conflict with security. Security should really win out here, therefore we recommend 1s.

```
kratos hashers argon2 calibrate [<desired-duration>] [flags]
```

### Options

```
      --adjust-memory-by string   amount by which the memory is adjusted in every step while probing (default "1GB")
  -h, --help                      help for calibrate
      --key-length uint32         length of the key in bytes (default 32)
      --max-memory string         maximum memory allowed (default no limit)
      --parallelism uint8         number of threads to use (default 72)
  -r, --probe-runs int            runs per probe, median of all runs is taken as the result (default 2)
      --salt-length uint32        length of the salt in bytes (default 16)
  -i, --start-iterations uint32   number of iterations to start probing at (default 1)
  -m, --start-memory string       amount of memory to start probing at (default "4GB")
  -v, --verbose                   verbose output
```

### Options inherited from parent commands

```
  -c, --config string   Path to config file. Supports .json, .yaml, .yml, .toml. Default is "$HOME/.kratos.(yaml|yml|toml|json)"
```

### SEE ALSO

- [kratos hashers argon2](kratos-hashers-argon2) -
