---
id: setting-up-password-hashing-parameters
title: Setting up Password Hashing Parameters
---

Currently, ORY Kratos supports password hashing using Argon2 in the Argon2id variant.
It is important to set up it's parameters to ensure a stable and reliable operation of ORY Kratos.
In essence, you want to fulfill the following constrains:

1. Duration: the execution time of one hashing operation - this translates to the response time of ORY Kratos on login and registration.
2. Memory: the amount of available memory on the server
3. Space: the amount of space for persistent storage

We recommend choosing a duration of 0.5s to 1s and as much memory as possible. To determine the exact
recommended values following security best practices, run the CLI helper included in ORY Kratos:

```
$ kratos hashers argon2 calibrate 1s
```

It will output the exact values to use in the [configuration](../reference/configuration.md).

Head to [our blogpost](https://ory.sh/argon2-parameter-choice-best-practice/) about Argon2 parameters
to learn how this command and password checking in ORY Kratos works.

If you encounter any problems like timeouts or out-of-memory errors, consolidate our
[troubleshooting guide](../debug/performance-out-of-memory-password-hashing-argon2.md).
