---
id: performance-out-of-memory-password-hashing-argon2
title: Performance Problems and Out Of Memory Panics Caused by Password Hashing
---

To securely check if passwords match, ORY Kratos stores the Argon2 hash of every
password. This algorithm has to be tuned to match the desired security level as
well as responsiveness. Because it is not easy to determine the exact values
without trying them out, ORY Kratos comes with a
[CLI](../cli/kratos-hashers-argon2-calibrate.md) that automatically calibrates
the values, following best practices. You can read more about these best
practices in our
[blog post](https://www.ory.sh/choose-recommended-argon2-parameters-password-hashing/).

## Common Errors

There are some errors that indicate your Argon2 parameters need adjustment:

1. Very slow login and registration requests, might cause network timeouts;
2. ORY Kratos fails with `fatal error: runtime: out of memory`;
3. The host environment (e.g. Minikube, Docker, Kubernetes...) crashes or
   becomes unresponsive;

In any of these cases, try reducing the resources used by Argon2 or increasing
the resources available to Kratos. Use the
[Argon2 calibrate CLI](../cli/kratos-hashers-argon2-calibrate.md) to detect the
best practice values for your server. Note that the calibration has to be done
under the exact same conditions that the server runs at.
