---
id: credentials
title: Overview
---

Each identity has one or more credentials associated with it:

```yaml
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
      - johnd@ory.sh
    config:
      hashed_password: ...
  oidc:
    id: oidc
    identifiers:
      - google:j8kf7a3...
      - facebook:83475891...
    config:
      - provider: google
        identifier: j8kf7a3
      - provider: facebook
        identifier: 83475891
```

Ory Kratos supports several credential types:

- `password`: The most common _identifier (username, email, ...) + password_
  credential.
- `oidc`: The "Log in with Google/Facebook/GitHub/..." credential.
- Other credentials - support other credential types (X509 Certificates,
  Biometrics, ...) at will be added a later stage.

Each credential - regardless of its type - has one or more identifiers attached
to it. Each identifier is universally unique. Assuming we had one identity with
credentials

```yaml
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
```

and tried to create (or update) another identity with the same identifier
(`john.doe@acme.com`), the system would reject the request with a 409 Conflict
state.

While credentials must be unique per type, there can be duplicates amongst
multiple types:

```yaml
# This is ok:
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
  oidc:
    id: oidc
    identifiers:
      - john.doe@acme.com
```

The same would apply if those were two separate identities:

```yaml
# Identity 1
credentials:
  password:
    id: password
    identifiers:
      - john.doe@acme.com
---
# Identity 2
credentials:
  oidc:
    id: oidc
    identifiers:
      - john.doe@acme.com
```
