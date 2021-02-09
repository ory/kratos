---
id: kratos-hashers-argon2-calibrate
title: kratos hashers argon2 calibrate
description: kratos hashers argon2 calibrate Computes Optimal Argon2 Parameters.
---

<!--
This file is auto-generated.

To improve this file please make your change against the appropriate "./cmd/*.go" file.
-->

## kratos hashers argon2 calibrate

Computes Optimal Argon2 Parameters.

### Synopsis

This command helps you calibrate the configuration parameters for Argon2.
Password hashing is a trade-off between security, resource consumption, and user
experience. Resource consumption should not be too high and the login should not
take too long.

We recommend that the login process takes between half a second and one second
for password hashing, giving a good balance between security and user
experience.

Please note that the values depend on the machine you run the hashing on. If you
have RAM constraints please choose lower memory targets to avoid out of memory
panics.

```
kratos hashers argon2 calibrate [<desired-duration>] [flags]
```

### Options

```
      --adjust-memory-by byte_size   Amount by which the memory is adjusted in every step while probing. (default 1.00GB)
  -h, --help                         help for calibrate
      --key-length uint32            Length of the key in bytes. (default 32)
      --max-memory byte_size         Maximum memory allowed (default no limit). (default 0.00B)
      --parallelism uint8            Number of threads to use. (default 16)
  -r, --probe-runs int               Runs per probe, median of all runs is taken as the result. (default 2)
  -q, --quiet                        Quiet output.
      --salt-length uint32           Length of the salt in bytes. (default 16)
  -i, --start-iterations uint32      Number of iterations to start probing at. (default 1)
  -m, --start-memory byte_size       Amount of memory to start probing at. (default 4.00GB)
```

### SEE ALSO

- [kratos hashers argon2](kratos-hashers-argon2) -
