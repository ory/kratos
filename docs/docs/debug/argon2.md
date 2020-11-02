---
id: argon2
title: Performance problems and crashes caused by Argon2
---

To securely check if passwords match, ORY Kratos stores the Argon2 hash of every password.
This algorithm has to be tuned to match the desired security level as well as responsiveness.
Because it is not easy to determine the exact values without trying them out, ORY Kratos
comes with a [CLI](../cli/kratos-hashers-argon2-calibrate.md) that automatically calibrates the values, following best practices.
You can read more about these best practices in our [blog post](todo).

## Common Errors

There are some errors that indicate your Argon2 parameters need adjustment:

1. very slow login and registration request, might cause network timeouts
2. Kratos fails with `fatal error: runtime: out of memory`
3. Kratos' environment (e.g. Minikube, Docker, Kubernetes...) crashes or gets unresponsive

In any of these cases, try reducing the resources used by Argon2 or increasing the resources available to Kratos.
Use the [Argon2 calibrate CLI](../cli/kratos-hashers-argon2-calibrate.md) to detect the best practice values for your server.
Note that the calibration has to be done under the exact same conditions that the server runs at.
